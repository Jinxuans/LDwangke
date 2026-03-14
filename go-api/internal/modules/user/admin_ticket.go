package user

import (
	"fmt"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) AdminTicketList(page, limit int, status, uid int, search string) ([]model.Ticket, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if status > 0 {
		where += " AND t.status = ?"
		args = append(args, status)
	}
	if uid > 0 {
		where += " AND t.uid = ?"
		args = append(args, uid)
	}
	if search != "" {
		where += " AND (t.content LIKE ? OR t.reply LIKE ? OR CAST(t.oid AS CHAR) LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket t WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT t.id, t.uid, COALESCE(t.oid,0), COALESCE(t.type,''), COALESCE(t.content,''), COALESCE(t.reply,''),
			t.status, COALESCE(DATE_FORMAT(t.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(t.reply_time,'%%Y-%%m-%%d %%H:%%i:%%s'),''),
			COALESCE(t.supplier_report_id,0), COALESCE(t.supplier_status,-1), COALESCE(t.supplier_answer,''),
			COALESCE(o.user,''), COALESCE(o.ptname,''), COALESCE(o.status,''), COALESCE(o.yid,''),
			COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		FROM qingka_wangke_ticket t
		LEFT JOIN qingka_wangke_order o ON o.oid = t.oid
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		LEFT JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		WHERE %s ORDER BY t.id DESC LIMIT ? OFFSET ?`, where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var ticket model.Ticket
		rows.Scan(&ticket.ID, &ticket.UID, &ticket.OID, &ticket.Type, &ticket.Content, &ticket.Reply,
			&ticket.Status, &ticket.AddTime, &ticket.ReplyTime,
			&ticket.SupplierReportID, &ticket.SupplierStatus, &ticket.SupplierAnswer,
			&ticket.OrderUser, &ticket.OrderPT, &ticket.OrderStatus, &ticket.OrderYID,
			&ticket.SupplierReportSwitch, &ticket.SupplierReportHID)
		tickets = append(tickets, ticket)
	}
	if tickets == nil {
		tickets = []model.Ticket{}
	}
	return tickets, total, nil
}

func (s *Service) TicketStats() (map[string]int64, error) {
	var total, pending, replied, closed, upPending int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket").Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 1").Scan(&pending)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 2").Scan(&replied)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 3").Scan(&closed)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE supplier_report_id > 0 AND supplier_status IN (0,4)").Scan(&upPending)
	return map[string]int64{
		"total": total, "pending": pending, "replied": replied,
		"closed": closed, "upstream_pending": upPending,
	}, nil
}

func (s *Service) AutoCloseExpiredTickets(days int) (int64, error) {
	result, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET status = 3 WHERE status = 2 AND reply_time IS NOT NULL AND reply_time < DATE_SUB(NOW(), INTERVAL ? DAY)",
		days,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Service) UpdateTicketSupplierReport(ticketID, reportID, supplierStatus int, answer string) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET supplier_report_id = ?, supplier_status = ?, supplier_answer = ? WHERE id = ?",
		reportID, supplierStatus, answer, ticketID,
	)
	return err
}

func (s *Service) GetTicketByID(ticketID int) (*model.Ticket, error) {
	var ticket model.Ticket
	err := database.DB.QueryRow(
		`SELECT t.id, t.uid, COALESCE(t.oid,0), COALESCE(t.type,''), COALESCE(t.content,''), COALESCE(t.reply,''),
			t.status, COALESCE(DATE_FORMAT(t.addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(t.reply_time,'%Y-%m-%d %H:%i:%s'),''),
			COALESCE(t.supplier_report_id,0), COALESCE(t.supplier_status,-1), COALESCE(t.supplier_answer,''),
			COALESCE(o.user,''), COALESCE(o.ptname,''), COALESCE(o.status,''), COALESCE(o.yid,'')
		FROM qingka_wangke_ticket t
		LEFT JOIN qingka_wangke_order o ON o.oid = t.oid
		WHERE t.id = ?`, ticketID,
	).Scan(&ticket.ID, &ticket.UID, &ticket.OID, &ticket.Type, &ticket.Content, &ticket.Reply,
		&ticket.Status, &ticket.AddTime, &ticket.ReplyTime,
		&ticket.SupplierReportID, &ticket.SupplierStatus, &ticket.SupplierAnswer,
		&ticket.OrderUser, &ticket.OrderPT, &ticket.OrderStatus, &ticket.OrderYID)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}
