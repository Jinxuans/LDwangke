package xm

import (
	"fmt"
	"strings"

	"go-api/internal/database"
)

// XMProjectAdmin 管理员视角的小米运动项目。
// 连接信息来自 provider，项目只保留本地可运营字段和上游映射字段。
type XMProjectAdmin struct {
	ID            int     `json:"id"`
	ProviderID    int     `json:"provider_id"`
	ProviderName  string  `json:"provider_name"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	UpstreamPrice float64 `json:"upstream_price"`
	Query         int     `json:"query"`
	Password      int     `json:"password"`
	URL           string  `json:"url"`
	UID           string  `json:"uid"`
	Key           string  `json:"key"`
	Token         string  `json:"token"`
	Type          int     `json:"type"`
	PID           string  `json:"p_id"`
	Status        int     `json:"status"`
	SortOrder     int     `json:"sort_order"`
	SyncMode      int     `json:"sync_mode"`
}

func (s *XMService) AdminListProjects() ([]XMProjectAdmin, error) {
	rows, err := database.DB.Query(`
		SELECT x.id,
		       COALESCE(x.provider_id, 0),
		       COALESCE(p.name, ''),
		       x.name,
		       COALESCE(x.description, ''),
		       COALESCE(NULLIF(x.local_price, 0), x.price, x.upstream_price, 0) AS local_price,
		       COALESCE(x.upstream_price, 0),
		       x.`+"`query`"+`,
		       x.password,
		       COALESCE(p.base_url, ''),
		       COALESCE(p.uid, ''),
		       COALESCE(p.`+"`key`"+`, ''),
		       COALESCE(p.token, ''),
		       COALESCE(p.auth_type, 0),
		       COALESCE(x.p_id, ''),
		       x.status,
		       COALESCE(x.sort_order, 0),
		       COALESCE(x.sync_mode, 1)
		FROM xm_project x
		LEFT JOIN xm_provider p ON p.id = x.provider_id
		WHERE x.is_deleted = 0
		ORDER BY x.sort_order ASC, x.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]XMProjectAdmin, 0)
	for rows.Next() {
		var item XMProjectAdmin
		if err := rows.Scan(
			&item.ID,
			&item.ProviderID,
			&item.ProviderName,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.UpstreamPrice,
			&item.Query,
			&item.Password,
			&item.URL,
			&item.UID,
			&item.Key,
			&item.Token,
			&item.Type,
			&item.PID,
			&item.Status,
			&item.SortOrder,
			&item.SyncMode,
		); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

// AdminSaveProject 只保存项目本地字段；连接字段仅用于兼容旧数据。
func (s *XMService) AdminSaveProject(p XMProjectAdmin) (int, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return 0, fmt.Errorf("项目名称不能为空")
	}
	if p.ProviderID <= 0 {
		return 0, fmt.Errorf("项目必须绑定上游连接")
	}
	if p.SortOrder <= 0 {
		p.SortOrder = p.ID
	}

	if p.ID > 0 {
		// 运行期只依赖 provider_id；连接字段由 provider 表统一维护。
		if existing, err := s.getProjectRow(p.ID); err == nil {
			if p.UpstreamPrice <= 0 {
				p.UpstreamPrice = parseXMFloat(existing["upstream_price"])
			}
			if p.Price <= 0 {
				p.Price = getXMProjectBasePrice(existing)
			}
		}

		_, err := database.DB.Exec(
			`UPDATE xm_project
			 SET provider_id=?, name=?, description=?, price=?, local_price=?, upstream_price=?, `+"`query`"+`=?, password=?, p_id=?, status=?, sort_order=?, sync_mode=?, updated_at=NOW()
			 WHERE id=?`,
			p.ProviderID,
			name,
			p.Description,
			p.Price,
			p.Price,
			p.UpstreamPrice,
			p.Query,
			p.Password,
			p.PID,
			p.Status,
			p.SortOrder,
			p.SyncMode,
			p.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存失败: %v", err)
		}
		return p.ID, nil
	}

	result, err := database.DB.Exec(
		`INSERT INTO xm_project
		 (provider_id, name, description, price, local_price, upstream_price, `+"`query`"+`, password, p_id, status, sort_order, sync_mode, is_deleted, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW())`,
		p.ProviderID,
		name,
		p.Description,
		p.Price,
		p.Price,
		p.UpstreamPrice,
		p.Query,
		p.Password,
		p.PID,
		p.Status,
		p.SortOrder,
		p.SyncMode,
	)
	if err != nil {
		return 0, fmt.Errorf("添加失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (s *XMService) AdminDeleteProject(id int) error {
	_, err := database.DB.Exec("UPDATE xm_project SET is_deleted = 1, updated_at = NOW() WHERE id = ?", id)
	return err
}
