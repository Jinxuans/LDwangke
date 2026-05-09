package sxgz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
)

type SxgzService struct {
	client *http.Client
}

var sxgzServiceInstance = &SxgzService{
	client: &http.Client{Timeout: 30 * time.Second},
}

var sxgzCompaniesRefreshMu sync.Mutex

func Sxgz() *SxgzService {
	return sxgzServiceInstance
}

type OrderQuoteRequest struct {
	ServiceType              string                 `json:"service_type"`
	CompanyID                int                    `json:"company_id"`
	CustomCompanyName        string                 `json:"custom_company_name"`
	CustomerName             string                 `json:"customer_name"`
	CustomerEmail            string                 `json:"customer_email"`
	CustomerPhone            string                 `json:"customer_phone"`
	CustomerAddress          string                 `json:"customer_address"`
	CourierCompany           string                 `json:"courier_company"`
	TrackingNumber           string                 `json:"tracking_number"`
	ReturnTrackingNumber     string                 `json:"return_tracking_number"`
	PrintCopies              int                    `json:"print_copies"`
	PrintOptions             []string               `json:"print_options"`
	FilePrintOptions         []SxgzFilePrintRequest `json:"file_print_options"`
	PaperSize                string                 `json:"paper_size"`
	SpecialRequirements      string                 `json:"special_requirements"`
	BusinessLicense          bool                   `json:"business_license"`
	OnlyBusinessLicense      bool                   `json:"only_business_license"`
	MaterialType             string                 `json:"material_type"`
	DeliveryOption           string                 `json:"delivery_option"`
	SelectedLicenseCompanies []int                  `json:"selected_license_companies"`
}

type SxgzOrder struct {
	OrderID              int64          `json:"order_id"`
	UID                  int            `json:"uid"`
	OrderNo              string         `json:"order_no"`
	ServiceType          string         `json:"service_type"`
	CompanyID            int            `json:"company_id"`
	CompanyName          string         `json:"company_name"`
	CustomCompanyName    sql.NullString `json:"-"`
	BusinessLicense      int            `json:"business_license"`
	OnlyBusinessLicense  int            `json:"only_business_license"`
	MaterialType         sql.NullString `json:"-"`
	UploadedFile         sql.NullString `json:"-"`
	OriginalFilename     sql.NullString `json:"-"`
	FileSize             sql.NullInt64  `json:"-"`
	CustomerName         string         `json:"customer_name"`
	CustomerEmail        sql.NullString `json:"-"`
	CustomerPhone        sql.NullString `json:"-"`
	CustomerAddress      sql.NullString `json:"-"`
	CourierCompany       sql.NullString `json:"-"`
	TrackingNumber       sql.NullString `json:"-"`
	ReturnTrackingNumber sql.NullString `json:"-"`
	PrintCopies          int            `json:"print_copies"`
	PrintOptions         sql.NullString `json:"-"`
	PaperSize            string         `json:"paper_size"`
	SpecialRequirements  sql.NullString `json:"-"`
	BasePrice            float64        `json:"base_price"`
	MailPrice            float64        `json:"mail_price"`
	PrintPrice           float64        `json:"print_price"`
	LicensePrice         float64        `json:"license_price"`
	TotalPrice           float64        `json:"total_price"`
	Status               string         `json:"status"`
	AdminNotes           sql.NullString `json:"-"`
	RefundReason         sql.NullString `json:"-"`
	ProcessedFiles       sql.NullString `json:"-"`
	ProcessedFileURL     sql.NullString `json:"-"`
	Source               string         `json:"source"`
	AgentUID             sql.NullInt64  `json:"-"`
	AgentOrderID         sql.NullInt64  `json:"-"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            sql.NullString `json:"-"`
	CompletedAt          sql.NullString `json:"-"`
}

type SxgzFilePrintRequest struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	PageCount  int    `json:"page_count"`
	PrintCount int    `json:"print_count"`
	PrintMode  string `json:"print_mode"`
	ColorMode  string `json:"color_mode"`
	PaperSize  string `json:"paper_size"`
	StampType  string `json:"stamp_type"`
}

type SxgzFilePrintOptions struct {
	PrintCount int    `json:"print_count,omitempty"`
	PrintMode  string `json:"print_mode,omitempty"`
	ColorMode  string `json:"color_mode,omitempty"`
	PaperSize  string `json:"paper_size,omitempty"`
	StampType  string `json:"stamp_type,omitempty"`
}

type SxgzOrderListResult struct {
	List  []map[string]any `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

type SxgzQuoteResult struct {
	BasePrice         float64 `json:"base_price"`
	PrintPrice        float64 `json:"print_price"`
	LicensePrice      float64 `json:"license_price"`
	ExtraOptionsPrice float64 `json:"extra_options_price"`
	TotalPrice        float64 `json:"total_price"`
	UserRate          float64 `json:"user_rate"`
	PriceMultiplier   float64 `json:"price_multiplier"`
	CompanyName       string  `json:"company_name"`
}

type SxgzCreateResult struct {
	OrderID    int64   `json:"order_id"`
	OrderNo    string  `json:"order_no"`
	UpstreamID int64   `json:"upstream_id,omitempty"`
	NeedRefund bool    `json:"need_refund,omitempty"`
	Message    string  `json:"message"`
	TotalPrice float64 `json:"total_price"`
}

type SxgzAnnouncementRequest struct {
	Page     int
	PageSize int
	Type     string
}

func (s *SxgzService) loadConfig() (SxgzConfig, error) {
	var raw string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", "sxgz_config").Scan(&raw)
	if err != nil {
		return defaultSxgzConfig(), nil
	}
	cfg, err := parseSxgzConfig(raw)
	if err != nil {
		return defaultSxgzConfig(), nil
	}
	return cfg, nil
}

func (s *SxgzService) saveConfig(cfg SxgzConfig) error {
	raw, err := cfg.Marshal()
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?) ON DUPLICATE KEY UPDATE k = VALUES(k)",
		"sxgz_config", raw,
	)
	return err
}

func (s *SxgzService) getUserRate(uid int) (float64, error) {
	var rate float64
	if err := database.DB.QueryRow("SELECT COALESCE(addprice, 1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate); err != nil {
		return 0, err
	}
	if rate <= 0 {
		rate = 1
	}
	return rate, nil
}

func (s *SxgzService) getCompanyByID(companyID int, allowRefresh bool) (*SxgzCompany, error) {
	var company SxgzCompany
	err := database.DB.QueryRow(
		"SELECT cid, name, COALESCE(price, 0), COALESCE(license_price, 0), COALESCE(content, ''), COALESCE(status, 0), COALESCE(raw_json, ''), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s'), COALESCE(source, 'upstream') FROM fd_sxgz_company_cache WHERE cid = ? LIMIT 1",
		companyID,
	).Scan(&company.CID, &company.Name, &company.Price, &company.LicensePrice, &company.Content, &company.Status, &company.RawJSON, &company.UpdatedAt, &company.Source)
	if err == nil {
		return &company, nil
	}
	if allowRefresh {
		if _, refreshErr := s.RefreshCompanies(context.Background()); refreshErr == nil {
			return s.getCompanyByID(companyID, false)
		}
	}
	return nil, fmt.Errorf("company not found")
}

func (s *SxgzService) listCompanies(search string, onlyLicense bool) ([]SxgzCompany, error) {
	where := "1=1"
	args := make([]any, 0, 2)
	if search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	if onlyLicense {
		where += " AND (name LIKE ? OR license_price > 0)"
		args = append(args, "%营业执照%")
	}

	rows, err := database.DB.Query(
		"SELECT cid, name, COALESCE(price,0), COALESCE(license_price,0), COALESCE(content,''), COALESCE(status,0), COALESCE(raw_json,''), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s'), COALESCE(source,'upstream') FROM fd_sxgz_company_cache WHERE "+where+" ORDER BY status DESC, cid ASC",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SxgzCompany, 0)
	for rows.Next() {
		var c SxgzCompany
		if err := rows.Scan(&c.CID, &c.Name, &c.Price, &c.LicensePrice, &c.Content, &c.Status, &c.RawJSON, &c.UpdatedAt, &c.Source); err != nil {
			continue
		}
		out = append(out, c)
	}
	if out == nil {
		out = []SxgzCompany{}
	}
	return out, nil
}

func (s *SxgzService) upsertCompanies(companies []SxgzCompany) error {
	for _, company := range companies {
		raw, _ := json.Marshal(company)
		_, err := database.DB.Exec(
			`INSERT INTO fd_sxgz_company_cache (cid, name, price, license_price, content, status, raw_json, source, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
			 ON DUPLICATE KEY UPDATE
			   name = VALUES(name),
			   price = VALUES(price),
			   license_price = VALUES(license_price),
			   content = VALUES(content),
			   status = VALUES(status),
			   raw_json = VALUES(raw_json),
			   source = VALUES(source),
			   updated_at = VALUES(updated_at)`,
			company.CID, company.Name, company.Price, company.LicensePrice, company.Content, company.Status, string(raw), company.Source,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyConfigMultiplierToCompanies(companies []SxgzCompany, multiplier float64) []SxgzCompany {
	if multiplier <= 0 {
		multiplier = 1
	}
	out := make([]SxgzCompany, len(companies))
	copy(out, companies)
	for i := range out {
		out[i].Price = round2(out[i].Price * multiplier)
		out[i].LicensePrice = round2(out[i].LicensePrice * multiplier)
	}
	return out
}

func (s *SxgzService) applyUserRateToCompanies(uid int, companies []SxgzCompany) []SxgzCompany {
	userRate, err := s.getUserRate(uid)
	if err != nil || userRate <= 0 {
		userRate = 1
	}
	out := make([]SxgzCompany, len(companies))
	copy(out, companies)
	for i := range out {
		out[i].Price = round2(out[i].Price * userRate)
		out[i].LicensePrice = round2(out[i].LicensePrice * userRate)
	}
	return out
}

func (s *SxgzService) RefreshCompanies(ctx context.Context) ([]SxgzCompany, error) {
	sxgzCompaniesRefreshMu.Lock()
	defer sxgzCompaniesRefreshMu.Unlock()

	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.UpstreamEnabled() {
		return nil, fmt.Errorf("upstream is not configured")
	}

	body, status, err := s.upstreamRequest(ctx, cfg, "get_companies_for_agent", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", status)
	}
	if v, ok := body["success"]; ok && !asBool(v) {
		msg := strings.TrimSpace(asString(body["message"]))
		if msg == "" {
			msg = "upstream returned failure"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	companies := decodeCompaniesPayload(body)
	for i := range companies {
		companies[i].Source = "upstream"
	}
	if len(companies) == 0 {
		return nil, fmt.Errorf("upstream returned empty company list")
	}
	companies = applyConfigMultiplierToCompanies(companies, cfg.PriceMultiplier)
	if err := s.upsertCompanies(companies); err != nil {
		return nil, err
	}
	return companies, nil
}

func (s *SxgzService) GetAnnouncements(ctx context.Context, req SxgzAnnouncementRequest) (*SxgzAnnouncementListResult, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.UpstreamEnabled() {
		return nil, fmt.Errorf("upstream is not configured")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if strings.TrimSpace(req.Type) == "" {
		req.Type = "全站公告"
	}

	body, status, err := s.upstreamRequest(ctx, cfg, "get_gonggao", http.MethodGet, map[string]any{
		"page":     req.Page,
		"pageSize": req.PageSize,
		"type":     req.Type,
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", status)
	}
	if v, ok := body["success"]; ok && !asBool(v) {
		msg := strings.TrimSpace(asString(body["message"]))
		if msg == "" {
			msg = strings.TrimSpace(asString(body["msg"]))
		}
		if msg == "" {
			msg = "upstream returned failure"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if v, ok := body["code"]; ok {
		code := asInt(v)
		if code != 0 && code != 1 {
			msg := strings.TrimSpace(asString(body["msg"]))
			if msg == "" {
				msg = strings.TrimSpace(asString(body["message"]))
			}
			if msg == "" {
				msg = "upstream returned failure"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	result := &SxgzAnnouncementListResult{
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     req.Type,
	}
	if v, ok := body["total"]; ok {
		result.Total = asInt(v)
	}
	if v, ok := body["hasMore"]; ok {
		result.HasMore = asBool(v)
	}
	result.Data = decodeAnnouncementsPayload(body)
	if result.Total == 0 {
		result.Total = len(result.Data)
	}
	if !result.HasMore && result.Total > 0 {
		result.HasMore = result.Page*result.PageSize < result.Total
	}
	return result, nil
}

func decodeAnnouncementsPayload(payload map[string]any) []SxgzAnnouncement {
	raw := any(nil)
	if v, ok := payload["data"]; ok {
		raw = v
	} else if v, ok := payload["list"]; ok {
		raw = v
	}
	switch val := raw.(type) {
	case []any:
		return decodeAnnouncementsSlice(val)
	case map[string]any:
		if nested, ok := val["list"].([]any); ok {
			return decodeAnnouncementsSlice(nested)
		}
		if nested, ok := val["data"].([]any); ok {
			return decodeAnnouncementsSlice(nested)
		}
	}
	return []SxgzAnnouncement{}
}

func decodeAnnouncementsSlice(items []any) []SxgzAnnouncement {
	out := make([]SxgzAnnouncement, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		announcement := SxgzAnnouncement{
			AID:         asInt(row["AID"]),
			Title:       asString(row["Title"]),
			Content:     asString(row["Content"]),
			PublishDate: asString(row["PublishDate"]),
			Importance:  asInt(row["Importance"]),
		}
		if announcement.AID == 0 {
			announcement.AID = asInt(row["id"])
		}
		if announcement.Title == "" {
			announcement.Title = asString(row["title"])
		}
		if announcement.Content == "" {
			announcement.Content = asString(row["content"])
		}
		if announcement.PublishDate == "" {
			announcement.PublishDate = asString(row["time"])
		}
		if announcement.Importance == 0 {
			announcement.Importance = asInt(row["importance"])
		}
		out = append(out, announcement)
	}
	return out
}

func decodeCompaniesPayload(payload map[string]any) []SxgzCompany {
	raw := any(nil)
	if v, ok := payload["data"]; ok {
		raw = v
	} else if v, ok := payload["companies"]; ok {
		raw = v
	}
	switch val := raw.(type) {
	case []any:
		return decodeCompaniesSlice(val)
	case map[string]any:
		if nested, ok := val["companies"].([]any); ok {
			return decodeCompaniesSlice(nested)
		}
	}
	return []SxgzCompany{}
}

func decodeCompaniesSlice(items []any) []SxgzCompany {
	out := make([]SxgzCompany, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		company := SxgzCompany{
			CID:          asInt(row["cid"]),
			Name:         asString(row["name"]),
			Price:        asFloat(row["price"]),
			LicensePrice: asFloat(row["license_price"]),
			Content:      asString(row["content"]),
			Status:       asInt(row["status"]),
			Source:       asString(row["source"]),
		}
		out = append(out, company)
	}
	return out
}

func (s *SxgzService) GetCompanies(uid int, search string) ([]SxgzCompany, error) {
	list, err := s.listCompanies(search, false)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		if _, refreshErr := s.RefreshCompanies(context.Background()); refreshErr == nil {
			list, err = s.listCompanies(search, false)
			if err != nil {
				return nil, err
			}
		}
	}
	return s.applyUserRateToCompanies(uid, list), nil
}

func (s *SxgzService) GetLicenseCompanies(uid int, search string) ([]SxgzCompany, error) {
	list, err := s.listCompanies(search, true)
	if err != nil {
		return nil, err
	}
	return s.applyUserRateToCompanies(uid, list), nil
}

func (s *SxgzService) Quote(uid int, req OrderQuoteRequest) (*SxgzQuoteResult, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	company, err := s.getCompanyByID(req.CompanyID, true)
	if err != nil {
		return nil, fmt.Errorf("company not found")
	}

	userRate, err := s.getUserRate(uid)
	if err != nil {
		return nil, err
	}
	if userRate <= 0 {
		userRate = 1
	}
	priceMultiplier := cfg.PriceMultiplier
	if priceMultiplier <= 0 {
		priceMultiplier = 1
	}

	basePrice := company.Price * userRate
	printPrice := 0.0
	effectivePrintCopies := req.PrintCopies
	if len(req.FilePrintOptions) > 0 {
		effectivePrintCopies = totalFilePrintSheets(req.FilePrintOptions)
	}
	if effectivePrintCopies > 0 {
		if effectivePrintCopies > cfg.PrintPricing.BaseFreeCopies {
			printPrice = float64(effectivePrintCopies-cfg.PrintPricing.BaseFreeCopies) * cfg.PrintPricing.ExtraCopyPrice * priceMultiplier
		}
	}
	licensePrice := 0.0
	if req.BusinessLicense {
		licensePrice = s.calcLicensePrice(cfg, req.SelectedLicenseCompanies, company, userRate)
	}
	if req.OnlyBusinessLicense {
		basePrice = 0
		printPrice = 0
	}

	extraOptions := 0.0
	printOptions := req.PrintOptions
	if len(req.FilePrintOptions) > 0 && len(printOptions) == 0 {
		printOptions = collectFilePrintOptionNames(req.FilePrintOptions)
	}
	for _, option := range printOptions {
		if rule, ok := cfg.PrintOptions[option]; ok {
			price := rule.Price * priceMultiplier
			if rule.AffectedByUserRate {
				price *= userRate
			}
			extraOptions += price
		}
	}
	if rule, ok := cfg.DeliveryOptions[req.DeliveryOption]; ok {
		price := rule.Price * priceMultiplier
		if rule.AffectedByUserRate {
			price *= userRate
		}
		extraOptions += price
	}

	total := basePrice + printPrice + licensePrice + extraOptions
	total = math.Round(total*100) / 100

	return &SxgzQuoteResult{
		BasePrice:         round2(basePrice),
		PrintPrice:        round2(printPrice),
		LicensePrice:      round2(licensePrice),
		ExtraOptionsPrice: round2(extraOptions),
		TotalPrice:        total,
		UserRate:          userRate,
		PriceMultiplier:   priceMultiplier,
		CompanyName:       company.Name,
	}, nil
}

func (s *SxgzService) calcLicensePrice(_ SxgzConfig, selectedIDs []int, company *SxgzCompany, userRate float64) float64 {
	if len(selectedIDs) > 0 {
		total := 0.0
		for _, companyID := range selectedIDs {
			row, err := s.getCompanyByID(companyID, false)
			if err != nil {
				continue
			}
			price := row.LicensePrice
			if price <= 0 {
				price = row.Price
			}
			total += price * userRate
		}
		if total > 0 {
			return total
		}
	}

	licenseCompanies, err := s.listCompanies("", true)
	if err == nil && len(licenseCompanies) > 0 {
		price := licenseCompanies[0].LicensePrice
		if price <= 0 {
			price = licenseCompanies[0].Price
		}
		if price > 0 {
			return price * userRate
		}
	}
	if company != nil && company.LicensePrice > 0 {
		return company.LicensePrice * userRate
	}
	return 0
}

func (s *SxgzService) CreateOrder(uid int, req OrderQuoteRequest, baseURL string) (*SxgzCreateResult, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	quote, err := s.Quote(uid, req)
	if err != nil {
		return nil, err
	}

	orderNo := fmt.Sprintf("SXGZ%s%06d", time.Now().Format("20060102150405"), uid%1000000)
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentMoney float64
	if err := tx.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ? FOR UPDATE", uid).Scan(&currentMoney); err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if currentMoney < quote.TotalPrice {
		return nil, fmt.Errorf("balance insufficient")
	}

	uploadedFiles := "[]"
	originalFilename := ""
	if req.MaterialType == "" {
		req.MaterialType = "upload"
	}
	effectivePrintCopies := req.PrintCopies
	if len(req.FilePrintOptions) > 0 {
		effectivePrintCopies = totalFilePrintSheets(req.FilePrintOptions)
	}
	effectivePrintOptions := req.PrintOptions
	if len(req.FilePrintOptions) > 0 && len(effectivePrintOptions) == 0 {
		effectivePrintOptions = collectFilePrintOptionNames(req.FilePrintOptions)
	}
	effectivePaperSize := normalizePaperSize(req.PaperSize)
	if effectivePaperSize == "A4" && len(req.FilePrintOptions) > 0 {
		effectivePaperSize = normalizePaperSize(req.FilePrintOptions[0].PaperSize)
	}

	result, err := tx.Exec(
		`INSERT INTO fd_sxgz_orders (
			uid, order_no, service_type, company_id, company_name, custom_company_name, business_license, only_business_license,
			material_type, uploaded_file, original_filename, customer_name, customer_email, customer_phone, customer_address,
			courier_company, tracking_number, return_tracking_number, print_copies, print_options, paper_size,
			special_requirements, base_price, mail_price, print_price, license_price, total_price, status, source,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 'direct', ?, ?)`,
		uid, orderNo, req.ServiceType, req.CompanyID, quote.CompanyName, strings.TrimSpace(req.CustomCompanyName), boolToInt(req.BusinessLicense), boolToInt(req.OnlyBusinessLicense),
		req.MaterialType, uploadedFiles, originalFilename, req.CustomerName, req.CustomerEmail, req.CustomerPhone, req.CustomerAddress,
		req.CourierCompany, req.TrackingNumber, req.ReturnTrackingNumber, effectivePrintCopies, mustJSON(effectivePrintOptions), effectivePaperSize,
		req.SpecialRequirements, quote.BasePrice, 0, quote.PrintPrice, quote.LicensePrice, quote.TotalPrice, now, now,
	)
	if err != nil {
		return nil, err
	}
	orderID, _ := result.LastInsertId()

	if quote.TotalPrice > 0 {
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", quote.TotalPrice, uid); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, ?, ?, ?, ?, ?)",
			uid, "扣费", -quote.TotalPrice, currentMoney-quote.TotalPrice, fmt.Sprintf("SXGZ order %s", orderNo), now,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	created := &SxgzCreateResult{OrderID: orderID, OrderNo: orderNo, Message: "order created", TotalPrice: quote.TotalPrice}

	if cfg.UpstreamEnabled() {
		upstreamPayload := map[string]any{
			"service_type":               req.ServiceType,
			"company_id":                 req.CompanyID,
			"company_name":               quote.CompanyName,
			"custom_company_name":        strings.TrimSpace(req.CustomCompanyName),
			"customer_name":              req.CustomerName,
			"customer_email":             req.CustomerEmail,
			"customer_phone":             req.CustomerPhone,
			"customer_address":           req.CustomerAddress,
			"courier_company":            req.CourierCompany,
			"tracking_number":            req.TrackingNumber,
			"return_tracking_number":     req.ReturnTrackingNumber,
			"print_copies":               effectivePrintCopies,
			"print_options":              effectivePrintOptions,
			"file_print_options":         req.FilePrintOptions,
			"paper_size":                 effectivePaperSize,
			"special_requirements":       req.SpecialRequirements,
			"business_license":           req.BusinessLicense,
			"only_business_license":      req.OnlyBusinessLicense,
			"material_type":              req.MaterialType,
			"delivery_option":            req.DeliveryOption,
			"selected_license_companies": req.SelectedLicenseCompanies,
		}
		upstreamResp, _, upstreamErr := s.upstreamRequest(context.Background(), cfg, "create_order", http.MethodPost, upstreamPayload)
		if upstreamErr != nil || !asBool(upstreamResp["success"]) {
			msg := "upstream create failed"
			if upstreamResp != nil {
				if v, ok := upstreamResp["message"].(string); ok && v != "" {
					msg = v
				}
			}
			_ = s.markFailedAndRefund(orderID, uid, quote.TotalPrice, msg)
			created.NeedRefund = true
			return created, fmt.Errorf("%s", msg)
		}
		if data, ok := upstreamResp["data"].(map[string]any); ok {
			if upstreamOrderNo := asString(data["order_no"]); upstreamOrderNo != "" {
				orderNo = upstreamOrderNo
			}
			var upstreamOrderID int64
			if v := asInt64(data["order_id"]); v > 0 {
				upstreamOrderID = v
			}
			_, _ = database.DB.Exec("UPDATE fd_sxgz_orders SET order_no = ?, agent_order_id = ?, updated_at = NOW() WHERE order_id = ?", orderNo, upstreamOrderID, orderID)
			created.OrderNo = orderNo
			created.UpstreamID = upstreamOrderID
		}
	}

	return created, nil
}

func (s *SxgzService) markFailedAndRefund(orderID int64, uid int, amount float64, reason string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var refunded float64
	err = tx.QueryRow("SELECT status, COALESCE(total_price,0) FROM fd_sxgz_orders WHERE order_id = ? FOR UPDATE", orderID).Scan(&status, &refunded)
	if err != nil {
		return err
	}
	if status == "refunded" {
		return tx.Commit()
	}
	_, err = tx.Exec("UPDATE fd_sxgz_orders SET status = 'failed', refund_reason = ?, updated_at = NOW() WHERE order_id = ?", reason, orderID)
	if err != nil {
		return err
	}
	if amount > 0 {
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", amount, uid); err != nil {
			return err
		}
		if _, err := tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, ?, ?, (SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?), ?, NOW())",
			uid, "退款", amount, uid, fmt.Sprintf("SXGZ refund: %s", reason),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SxgzService) GetOrder(uid int, orderID int64, isAdmin bool) (*SxgzOrder, error) {
	where := "order_id = ?"
	args := []any{orderID}
	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}

	row := database.DB.QueryRow(
		"SELECT order_id, uid, order_no, service_type, company_id, company_name, custom_company_name, business_license, only_business_license, material_type, uploaded_file, original_filename, file_size, customer_name, customer_email, customer_phone, customer_address, courier_company, tracking_number, return_tracking_number, print_copies, print_options, paper_size, special_requirements, base_price, mail_price, print_price, license_price, total_price, status, admin_notes, refund_reason, processed_files, processed_file_url, source, agent_uid, agent_order_id, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(completed_at, '%Y-%m-%d %H:%i:%s') FROM fd_sxgz_orders WHERE "+where+" LIMIT 1",
		args...,
	)
	return scanOrder(row)
}

func (s *SxgzService) ListOrders(uid int, isAdmin bool, page, size int, search, status string) (*SxgzOrderListResult, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	where := "1=1"
	args := make([]any, 0, 4)
	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}
	if search != "" {
		where += " AND (order_no LIKE ? OR customer_name LIKE ? OR company_name LIKE ?)"
		q := "%" + search + "%"
		args = append(args, q, q, q)
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM fd_sxgz_orders WHERE "+where, args...).Scan(&total); err != nil {
		return nil, err
	}

	query := "SELECT order_id, uid, order_no, service_type, company_id, company_name, custom_company_name, business_license, only_business_license, material_type, uploaded_file, original_filename, file_size, customer_name, customer_email, customer_phone, customer_address, courier_company, tracking_number, return_tracking_number, print_copies, print_options, paper_size, special_requirements, base_price, mail_price, print_price, license_price, total_price, status, admin_notes, refund_reason, processed_files, processed_file_url, source, agent_uid, agent_order_id, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(completed_at, '%Y-%m-%d %H:%i:%s') FROM fd_sxgz_orders WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			continue
		}
		out = append(out, orderToMap(order))
	}
	if out == nil {
		out = []map[string]any{}
	}

	return &SxgzOrderListResult{List: out, Total: total, Page: page, Size: size}, nil
}

func scanOrder(scanner interface{ Scan(...any) error }) (*SxgzOrder, error) {
	var order SxgzOrder
	err := scanner.Scan(
		&order.OrderID, &order.UID, &order.OrderNo, &order.ServiceType, &order.CompanyID, &order.CompanyName, &order.CustomCompanyName, &order.BusinessLicense, &order.OnlyBusinessLicense,
		&order.MaterialType, &order.UploadedFile, &order.OriginalFilename, &order.FileSize, &order.CustomerName, &order.CustomerEmail, &order.CustomerPhone,
		&order.CustomerAddress, &order.CourierCompany, &order.TrackingNumber, &order.ReturnTrackingNumber, &order.PrintCopies, &order.PrintOptions, &order.PaperSize,
		&order.SpecialRequirements, &order.BasePrice, &order.MailPrice, &order.PrintPrice, &order.LicensePrice, &order.TotalPrice, &order.Status, &order.AdminNotes,
		&order.RefundReason, &order.ProcessedFiles, &order.ProcessedFileURL, &order.Source, &order.AgentUID, &order.AgentOrderID, &order.CreatedAt, &order.UpdatedAt, &order.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func orderToMap(order *SxgzOrder) map[string]any {
	return map[string]any{
		"order_id":               order.OrderID,
		"uid":                    order.UID,
		"order_no":               order.OrderNo,
		"service_type":           order.ServiceType,
		"company_id":             order.CompanyID,
		"company_name":           order.CompanyName,
		"custom_company_name":    nullString(order.CustomCompanyName),
		"business_license":       order.BusinessLicense,
		"only_business_license":  order.OnlyBusinessLicense,
		"material_type":          nullString(order.MaterialType),
		"uploaded_file":          nullString(order.UploadedFile),
		"original_filename":      nullString(order.OriginalFilename),
		"file_size":              nullInt64(order.FileSize),
		"customer_name":          order.CustomerName,
		"customer_email":         nullString(order.CustomerEmail),
		"customer_phone":         nullString(order.CustomerPhone),
		"customer_address":       nullString(order.CustomerAddress),
		"courier_company":        nullString(order.CourierCompany),
		"tracking_number":        nullString(order.TrackingNumber),
		"return_tracking_number": nullString(order.ReturnTrackingNumber),
		"print_copies":           order.PrintCopies,
		"print_options":          nullString(order.PrintOptions),
		"paper_size":             order.PaperSize,
		"special_requirements":   nullString(order.SpecialRequirements),
		"base_price":             round2(order.BasePrice),
		"mail_price":             round2(order.MailPrice),
		"print_price":            round2(order.PrintPrice),
		"license_price":          round2(order.LicensePrice),
		"total_price":            round2(order.TotalPrice),
		"status":                 order.Status,
		"admin_notes":            nullString(order.AdminNotes),
		"refund_reason":          nullString(order.RefundReason),
		"processed_files":        nullString(order.ProcessedFiles),
		"processed_file_url":     nullString(order.ProcessedFileURL),
		"source":                 order.Source,
		"agent_uid":              nullInt64(order.AgentUID),
		"agent_order_id":         nullInt64(order.AgentOrderID),
		"created_at":             order.CreatedAt,
		"updated_at":             nullString(order.UpdatedAt),
		"completed_at":           nullString(order.CompletedAt),
		"files":                  parseOrderFiles(order.UploadedFile, order.ProcessedFileURL),
	}
}

func parseOrderFiles(uploaded sql.NullString, processed sql.NullString) map[string][]SxgzFileRecord {
	out := map[string][]SxgzFileRecord{
		"uploaded":  {},
		"processed": {},
	}
	if uploaded.Valid && strings.TrimSpace(uploaded.String) != "" {
		out["uploaded"] = parseFileRecords(uploaded.String)
	}
	if processed.Valid && strings.TrimSpace(processed.String) != "" {
		out["processed"] = parseFileRecords(processed.String)
	}
	return out
}

func parseFileRecords(raw string) []SxgzFileRecord {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SxgzFileRecord{}
	}
	var records []SxgzFileRecord
	if err := json.Unmarshal([]byte(raw), &records); err == nil {
		return records
	}
	var stringsOnly []string
	if err := json.Unmarshal([]byte(raw), &stringsOnly); err == nil {
		for _, item := range stringsOnly {
			records = append(records, SxgzFileRecord{URL: item, Name: filepath.Base(item), Storage: "remote"})
		}
		return records
	}
	return []SxgzFileRecord{{URL: raw, Name: filepath.Base(raw), Storage: "remote"}}
}

func (s *SxgzService) UpdateOrderStatus(uid int, orderID int64, status, notes, refundReason string, isAdmin bool) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var current SxgzOrder
	query := "SELECT order_id, uid, order_no, service_type, company_id, company_name, custom_company_name, business_license, only_business_license, material_type, uploaded_file, original_filename, file_size, customer_name, customer_email, customer_phone, customer_address, courier_company, tracking_number, return_tracking_number, print_copies, print_options, paper_size, special_requirements, base_price, mail_price, print_price, license_price, total_price, status, admin_notes, refund_reason, processed_files, processed_file_url, source, agent_uid, agent_order_id, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(completed_at, '%Y-%m-%d %H:%i:%s') FROM fd_sxgz_orders WHERE order_id = ?"
	args := []any{orderID}
	if !isAdmin {
		query += " AND uid = ?"
		args = append(args, uid)
	}
	row := tx.QueryRow(query, args...)
	if err := row.Scan(&current.OrderID, &current.UID, &current.OrderNo, &current.ServiceType, &current.CompanyID, &current.CompanyName, &current.CustomCompanyName, &current.BusinessLicense, &current.OnlyBusinessLicense, &current.MaterialType, &current.UploadedFile, &current.OriginalFilename, &current.FileSize, &current.CustomerName, &current.CustomerEmail, &current.CustomerPhone, &current.CustomerAddress, &current.CourierCompany, &current.TrackingNumber, &current.ReturnTrackingNumber, &current.PrintCopies, &current.PrintOptions, &current.PaperSize, &current.SpecialRequirements, &current.BasePrice, &current.MailPrice, &current.PrintPrice, &current.LicensePrice, &current.TotalPrice, &current.Status, &current.AdminNotes, &current.RefundReason, &current.ProcessedFiles, &current.ProcessedFileURL, &current.Source, &current.AgentUID, &current.AgentOrderID, &current.CreatedAt, &current.UpdatedAt, &current.CompletedAt); err != nil {
		return err
	}
	if !isAdmin && current.UID != uid {
		return fmt.Errorf("forbidden")
	}

	updateFields := []string{"status = ?", "updated_at = NOW()"}
	updateArgs := []any{status}
	if notes != "" {
		updateFields = append(updateFields, "admin_notes = ?")
		updateArgs = append(updateArgs, notes)
	}
	if refundReason != "" {
		updateFields = append(updateFields, "refund_reason = ?")
		updateArgs = append(updateArgs, refundReason)
	}
	if status == "completed" || status == "delivered" {
		updateFields = append(updateFields, "completed_at = NOW()")
	}
	queryArgs := append([]any{}, updateArgs...)
	queryArgs = append(queryArgs, orderID)
	updateSQL := "UPDATE fd_sxgz_orders SET " + strings.Join(updateFields, ", ") + " WHERE order_id = ?"
	if !isAdmin {
		updateSQL += " AND uid = ?"
		queryArgs = append(queryArgs, uid)
	}
	if _, err := tx.Exec(updateSQL, queryArgs...); err != nil {
		return err
	}

	if (status == "refunded" || status == "failed") && current.Status != status {
		amount := round2(current.TotalPrice)
		if amount > 0 {
			if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", amount, current.UID); err != nil {
				return err
			}
			_, _ = tx.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, ?, ?, (SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?), ?, NOW())",
				current.UID, "退款", amount, current.UID, fmt.Sprintf("SXGZ order %s refunded", current.OrderNo),
			)
		}
	}

	return tx.Commit()
}

func (s *SxgzService) ApplyRefund(uid int, orderID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("退款原因不能为空")
	}

	current, err := s.GetOrder(uid, orderID, false)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("订单不存在或无权限")
	}
	if current.Status == "cancelled" || current.Status == "refunded" || current.Status == "refund_requested" {
		return fmt.Errorf("该订单状态不允许申请退款")
	}

	cfg, err := s.loadConfig()
	if err == nil && cfg.UpstreamEnabled() {
		if _, upstreamErr := s.upstreamApplyRefund(context.Background(), cfg, current, reason); upstreamErr != nil {
			return upstreamErr
		}
	}

	return s.UpdateOrderStatus(uid, orderID, "refund_requested", "", reason, false)
}

func (s *SxgzService) SyncOrders(ctx context.Context) (int, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return 0, err
	}
	if !cfg.UpstreamEnabled() {
		return 0, fmt.Errorf("upstream is not configured")
	}

	resp, _, err := s.upstreamRequest(ctx, cfg, "sync_orders", http.MethodGet, nil)
	if err != nil {
		return 0, err
	}
	if v, ok := resp["success"]; ok && !asBool(v) {
		msg := strings.TrimSpace(asString(resp["message"]))
		if msg == "" {
			msg = "upstream returned failure"
		}
		return 0, fmt.Errorf("%s", msg)
	}

	items := decodeOrderSyncPayload(resp)
	updated := 0
	for _, item := range items {
		orderNo := asString(item["order_no"])
		if orderNo == "" {
			continue
		}
		tx, err := database.DB.Begin()
		if err != nil {
			continue
		}
		var orderID int64
		var currentUID int
		var currentStatus string
		var currentTotal float64
		if err := tx.QueryRow(
			"SELECT order_id, uid, status, COALESCE(total_price,0) FROM fd_sxgz_orders WHERE order_no = ? LIMIT 1 FOR UPDATE",
			orderNo,
		).Scan(&orderID, &currentUID, &currentStatus, &currentTotal); err != nil {
			_ = tx.Rollback()
			continue
		}
		fields := []string{"status = ?", "updated_at = NOW()"}
		args := []any{asString(item["status"])}
		if v := asString(item["admin_notes"]); v != "" {
			fields = append(fields, "admin_notes = ?")
			args = append(args, v)
		}
		if v := asString(item["completed_at"]); v != "" {
			fields = append(fields, "completed_at = ?")
			args = append(args, v)
		}
		if v := asString(item["processed_file_url"]); v != "" {
			fields = append(fields, "processed_file_url = ?")
			args = append(args, v)
		}
		args = append(args, orderID)
		if _, err := tx.Exec("UPDATE fd_sxgz_orders SET "+strings.Join(fields, ", ")+" WHERE order_id = ?", args...); err != nil {
			_ = tx.Rollback()
			continue
		}

		newStatus := asString(item["status"])
		if (newStatus == "refunded" || newStatus == "failed") && currentStatus != newStatus && currentTotal > 0 {
			if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", currentTotal, currentUID); err != nil {
				_ = tx.Rollback()
				continue
			}
			_, _ = tx.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, ?, ?, (SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?), ?, NOW())",
				currentUID, "退款", currentTotal, currentUID, fmt.Sprintf("SXGZ order %s refunded by sync", orderNo),
			)
		}
		if err := tx.Commit(); err == nil {
			updated++
		} else {
			_ = tx.Rollback()
		}
	}
	return updated, nil
}

func decodeOrderSyncPayload(payload map[string]any) []map[string]any {
	raw := any(nil)
	if v, ok := payload["data"]; ok {
		raw = v
	}
	switch val := raw.(type) {
	case []any:
		out := make([]map[string]any, 0, len(val))
		for _, item := range val {
			if row, ok := item.(map[string]any); ok {
				out = append(out, row)
			}
		}
		return out
	case map[string]any:
		if list, ok := val["list"].([]any); ok {
			out := make([]map[string]any, 0, len(list))
			for _, item := range list {
				if row, ok := item.(map[string]any); ok {
					out = append(out, row)
				}
			}
			return out
		}
	}
	return []map[string]any{}
}

func (s *SxgzService) UploadFile(uid int, orderID int64, fileHeader *multipart.FileHeader, baseURL string, rawPrintOptions string, pageCount int) (map[string]any, error) {
	order, err := s.GetOrder(uid, orderID, uid == 1)
	if err != nil {
		return nil, err
	}

	if fileHeader == nil {
		return nil, fmt.Errorf("file is required")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedSXGZFileExt(ext) {
		return nil, fmt.Errorf("unsupported file type")
	}
	if fileHeader.Size > 50*1024*1024 {
		return nil, fmt.Errorf("file too large")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	uploadDir := filepath.Join("uploads", "sxgz", fmt.Sprintf("uid_%d", uid), fmt.Sprintf("order_%d", orderID))
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, err
	}

	safeName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), orderID, ext)
	savePath := filepath.Join(uploadDir, safeName)
	dst, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return nil, err
	}
	if err := dst.Close(); err != nil {
		return nil, err
	}

	fileURL := "/" + filepath.ToSlash(savePath)
	printOptions := parseSxgzFilePrintOptions(rawPrintOptions)
	if pageCount <= 0 && printOptions != nil && printOptions.PrintCount > 0 {
		pageCount = 1
	}
	record := SxgzFileRecord{
		URL:          fileURL,
		Name:         fileHeader.Filename,
		Size:         fileHeader.Size,
		Storage:      "local",
		PageCount:    pageCount,
		PrintOptions: printOptions,
	}

	files := parseFileRecords(nullStringValue(order.UploadedFile))
	files = append(files, record)
	raw, _ := json.Marshal(files)
	allNames := make([]string, 0, len(files))
	totalSize := int64(0)
	for _, item := range files {
		allNames = append(allNames, item.Name)
		totalSize += item.Size
	}

	_, err = database.DB.Exec(
		"UPDATE fd_sxgz_orders SET uploaded_file = ?, original_filename = ?, file_size = ?, updated_at = NOW() WHERE order_id = ?",
		string(raw), strings.Join(allNames, ", "), totalSize, orderID,
	)
	if err != nil {
		return nil, err
	}

	if cfg, cfgErr := s.loadConfig(); cfgErr == nil && cfg.UpstreamEnabled() && order.AgentOrderID.Valid && order.AgentOrderID.Int64 > 0 {
		_, _ = s.callUpstreamUpdateFile(ctxWithTimeout(context.Background(), 30*time.Second), cfg, order, files, baseURL)
	}

	return map[string]any{
		"file_url":    fileURL,
		"file_name":   fileHeader.Filename,
		"size":        fileHeader.Size,
		"total_files": len(files),
	}, nil
}

func (s *SxgzService) ListFileRecords(uid int, orderID int64, isAdmin bool) (map[string][]SxgzFileRecord, error) {
	order, err := s.GetOrder(uid, orderID, isAdmin)
	if err != nil {
		return nil, err
	}
	return parseOrderFiles(order.UploadedFile, order.ProcessedFileURL), nil
}

func (s *SxgzService) AdminStats() (map[string]any, error) {
	stats := map[string]any{
		"overview": map[string]any{},
		"revenue":  map[string]any{},
	}
	counts := map[string]int64{}
	for _, status := range []string{"pending", "processing", "completed", "delivered", "cancelled", "failed", "refund_requested", "refunded"} {
		var count int64
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM fd_sxgz_orders WHERE status = ?", status).Scan(&count)
		counts[status] = count
	}
	var total, today, month float64
	_ = database.DB.QueryRow("SELECT COALESCE(SUM(total_price),0) FROM fd_sxgz_orders WHERE status IN ('completed','delivered')").Scan(&total)
	_ = database.DB.QueryRow("SELECT COALESCE(SUM(total_price),0) FROM fd_sxgz_orders WHERE status IN ('completed','delivered') AND DATE(created_at)=CURDATE()").Scan(&today)
	_ = database.DB.QueryRow("SELECT COALESCE(SUM(total_price),0) FROM fd_sxgz_orders WHERE status IN ('completed','delivered') AND YEAR(created_at)=YEAR(CURDATE()) AND MONTH(created_at)=MONTH(CURDATE())").Scan(&month)

	overview := map[string]any{
		"total_orders":      counts["pending"] + counts["processing"] + counts["completed"] + counts["delivered"] + counts["cancelled"] + counts["failed"] + counts["refund_requested"] + counts["refunded"],
		"pending_orders":    counts["pending"],
		"processing_orders": counts["processing"],
		"completed_orders":  counts["completed"] + counts["delivered"],
		"cancelled_orders":  counts["cancelled"],
		"failed_orders":     counts["failed"],
		"refund_orders":     counts["refund_requested"] + counts["refunded"],
	}
	var todayOrders int64
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM fd_sxgz_orders WHERE DATE(created_at) = CURDATE()").Scan(&todayOrders)
	overview["today_orders"] = todayOrders
	stats["overview"] = overview
	stats["revenue"] = map[string]any{
		"total": round2(total),
		"today": round2(today),
		"month": round2(month),
	}
	return stats, nil
}

func (s *SxgzService) callUpstreamUpdateFile(ctx context.Context, cfg SxgzConfig, order *SxgzOrder, files []SxgzFileRecord, baseURL string) (map[string]any, error) {
	payload := map[string]any{
		"action":               "update_order_file",
		"order_no":             order.OrderNo,
		"file_url":             mustJSON(files),
		"original_filename":    joinFileNames(files),
		"file_size":            totalFileSize(files),
		"plugin_domain":        baseURL,
		"special_requirements": nullString(order.SpecialRequirements),
	}
	resp, _, err := s.upstreamRequest(ctx, cfg, "update_order_file", http.MethodPost, payload)
	return resp, err
}

func joinFileNames(files []SxgzFileRecord) string {
	names := make([]string, 0, len(files))
	for _, item := range files {
		names = append(names, item.Name)
	}
	return strings.Join(names, ", ")
}

func totalFileSize(files []SxgzFileRecord) int64 {
	var sum int64
	for _, item := range files {
		sum += item.Size
	}
	return sum
}

func normalizeFilePrintRequest(item SxgzFilePrintRequest) SxgzFilePrintRequest {
	if item.PageCount <= 0 {
		item.PageCount = 1
	}
	if item.PrintCount <= 0 {
		item.PrintCount = 1
	}
	if strings.TrimSpace(item.PrintMode) == "" {
		item.PrintMode = "单面打印"
	}
	if strings.TrimSpace(item.ColorMode) == "" {
		item.ColorMode = "黑白"
	}
	if strings.TrimSpace(item.PaperSize) == "" {
		item.PaperSize = "A4纸"
	}
	if strings.TrimSpace(item.StampType) == "" {
		item.StampType = "实体章"
	}
	return item
}

func filePrintSheets(item SxgzFilePrintRequest) int {
	item = normalizeFilePrintRequest(item)
	sheetsPerCopy := item.PageCount
	if item.PrintMode == "双面打印" {
		sheetsPerCopy = (item.PageCount + 1) / 2
	}
	return sheetsPerCopy * item.PrintCount
}

func totalFilePrintSheets(items []SxgzFilePrintRequest) int {
	total := 0
	for _, item := range items {
		total += filePrintSheets(item)
	}
	return total
}

func collectFilePrintOptionNames(items []SxgzFilePrintRequest) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = normalizeFilePrintRequest(item)
		for _, option := range []string{item.PrintMode, item.ColorMode, item.PaperSize, item.StampType} {
			option = strings.TrimSpace(option)
			if option == "" || seen[option] {
				continue
			}
			seen[option] = true
			out = append(out, option)
		}
	}
	return out
}

func filePrintRequestOptions(item SxgzFilePrintRequest) *SxgzFilePrintOptions {
	item = normalizeFilePrintRequest(item)
	return &SxgzFilePrintOptions{
		PrintCount: item.PrintCount,
		PrintMode:  item.PrintMode,
		ColorMode:  item.ColorMode,
		PaperSize:  item.PaperSize,
		StampType:  item.StampType,
	}
}

func parseSxgzFilePrintOptions(raw string) *SxgzFilePrintOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item SxgzFilePrintRequest
	if err := json.Unmarshal([]byte(raw), &item); err == nil {
		if item.PrintCount > 0 || item.PrintMode != "" || item.ColorMode != "" || item.PaperSize != "" || item.StampType != "" {
			return filePrintRequestOptions(item)
		}
	}
	var legacy struct {
		PrintCount int    `json:"printCount"`
		PrintMode  string `json:"printMode"`
		ColorMode  string `json:"colorMode"`
		PaperSize  string `json:"paperSize"`
		StampType  string `json:"stampType"`
	}
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return nil
	}
	return filePrintRequestOptions(SxgzFilePrintRequest{
		PrintCount: legacy.PrintCount,
		PrintMode:  legacy.PrintMode,
		ColorMode:  legacy.ColorMode,
		PaperSize:  legacy.PaperSize,
		StampType:  legacy.StampType,
	})
}

func mustJSON(v any) string {
	if v == nil {
		return "[]"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func normalizePaperSize(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if strings.Contains(v, "A3") {
		return "A3"
	}
	if v != "A3" {
		return "A4"
	}
	return v
}

func isAllowedSXGZFileExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".pdf", ".doc", ".docx", ".zip", ".rar", ".7z", ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nullStringValue(v sql.NullString) string {
	return nullString(v)
}

func nullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func asString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case fmt.Stringer:
		return val.String()
	case float64:
		if math.Mod(val, 1) == 0 {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', 2, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case json.Number:
		return val.String()
	}
	return ""
}

func asInt(v any) int {
	return int(asInt64(v))
}

func asInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case json.Number:
		n, _ := val.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		return n
	}
	return 0
}

func asFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f
	}
	return 0
}

func asBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true") || val == "1"
	case float64:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	}
	return false
}

func ctxWithTimeout(parent context.Context, d time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, d)
	return ctx
}
