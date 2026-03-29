package xm

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"go-api/internal/database"
)

// XMProviderAdmin 表示一个可复用的上游连接配置。
// 一个 provider 对应一套认证参数，多个项目共享它。
type XMProviderAdmin struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	BaseURL      string  `json:"base_url"`
	AuthType     int     `json:"auth_type"`
	UID          string  `json:"uid"`
	Key          string  `json:"key"`
	Token        string  `json:"token"`
	Status       int     `json:"status"`
	Remark       string  `json:"remark"`
	LastSyncAt   string  `json:"last_sync_at"`
	ProjectCount int     `json:"project_count"`
}

// XMUpstreamProject 表示上游 getProjects 返回的一条项目数据。
type XMUpstreamProject struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Query       int     `json:"query"`
	Password    int     `json:"password"`
}

// XMImportProjectsRequest 表示批量导入上游项目时的选项。
type XMImportProjectsRequest struct {
	ProviderID          int      `json:"provider_id"`
	ProjectIDs          []string `json:"project_ids"`
	PriceMultiplier     float64  `json:"price_multiplier"`
	PriceAddition       float64  `json:"price_addition"`
	OverwriteLocalPrice bool     `json:"overwrite_local_price"`
}

// XMSyncProjectsRequest 表示一键同步上游项目时的选项。
type XMSyncProjectsRequest struct {
	ProviderID          int     `json:"provider_id"`
	SyncName            bool    `json:"sync_name"`
	SyncDescription     bool    `json:"sync_description"`
	SyncUpstreamPrice   bool    `json:"sync_upstream_price"`
	SyncQuery           bool    `json:"sync_query"`
	SyncPassword        bool    `json:"sync_password"`
	OverwriteLocalPrice bool    `json:"overwrite_local_price"`
	PriceMultiplier     float64 `json:"price_multiplier"`
	PriceAddition       float64 `json:"price_addition"`
}

type xmImportSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Total   int `json:"total"`
}

func roundXMPrice(price float64) float64 {
	return math.Round(price*10000) / 10000
}

// localPriceByRule 根据上游价格和导入策略，算出本地基础价。
// 这里不掺杂用户 addprice，保持和原系统的“项目基础价”一致。
func localPriceByRule(upstreamPrice float64, multiplier float64, addition float64) float64 {
	if multiplier <= 0 {
		multiplier = 1
	}
	return roundXMPrice(upstreamPrice*multiplier + addition)
}

func boolToTinyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func parseXMInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	case []uint8:
		s := strings.TrimSpace(string(t))
		if s == "" {
			return 0
		}
		var out int
		fmt.Sscanf(s, "%d", &out)
		return out
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0
		}
		var out int
		fmt.Sscanf(s, "%d", &out)
		return out
	default:
		return 0
	}
}

func parseXMFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float32:
		return float64(t)
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case []uint8:
		s := strings.TrimSpace(string(t))
		if s == "" {
			return 0
		}
		var out float64
		fmt.Sscanf(s, "%f", &out)
		return out
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0
		}
		var out float64
		fmt.Sscanf(s, "%f", &out)
		return out
	default:
		return 0
	}
}

func (s *XMService) getProviderRow(providerID int) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM xm_provider WHERE id = ? LIMIT 1", providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("上游连接不存在")
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)

	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}
	return result, nil
}

func (s *XMService) AdminListProviders() ([]XMProviderAdmin, error) {
	rows, err := database.DB.Query(`
		SELECT p.id, p.name, p.base_url, p.auth_type, p.uid, p.key, p.token, p.status,
		       COALESCE(p.remark, ''), COALESCE(DATE_FORMAT(p.last_sync_at, '%Y-%m-%d %H:%i:%s'), ''),
		       COUNT(x.id) AS project_count
		FROM xm_provider p
		LEFT JOIN xm_project x ON x.provider_id = p.id AND x.is_deleted = 0
		WHERE p.is_deleted = 0
		GROUP BY p.id
		ORDER BY p.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]XMProviderAdmin, 0)
	for rows.Next() {
		var item XMProviderAdmin
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.BaseURL,
			&item.AuthType,
			&item.UID,
			&item.Key,
			&item.Token,
			&item.Status,
			&item.Remark,
			&item.LastSyncAt,
			&item.ProjectCount,
		); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *XMService) AdminSaveProvider(p XMProviderAdmin) (int, error) {
	name := strings.TrimSpace(p.Name)
	baseURL := strings.TrimSpace(p.BaseURL)
	if name == "" {
		return 0, fmt.Errorf("连接名称不能为空")
	}
	if baseURL == "" {
		return 0, fmt.Errorf("API 地址不能为空")
	}
	if p.AuthType != 0 && p.AuthType != 1 {
		return 0, fmt.Errorf("认证类型无效")
	}
	if p.AuthType == 0 {
		if strings.TrimSpace(p.UID) == "" || strings.TrimSpace(p.Key) == "" {
			return 0, fmt.Errorf("Key 模式需要填写 UID 和 Key")
		}
	} else if strings.TrimSpace(p.Token) == "" {
		return 0, fmt.Errorf("Token 模式需要填写 Token")
	}

	if p.ID > 0 {
		_, err := database.DB.Exec(
			`UPDATE xm_provider
			 SET name=?, base_url=?, auth_type=?, uid=?, `+"`key`"+`=?, token=?, status=?, remark=?, updated_at=NOW()
			 WHERE id=?`,
			name, baseURL, p.AuthType, strings.TrimSpace(p.UID), strings.TrimSpace(p.Key), strings.TrimSpace(p.Token), p.Status, strings.TrimSpace(p.Remark), p.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存连接失败: %v", err)
		}
		return p.ID, nil
	}

	result, err := database.DB.Exec(
		`INSERT INTO xm_provider (name, base_url, auth_type, uid, `+"`key`"+`, token, status, remark, is_deleted, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW())`,
		name, baseURL, p.AuthType, strings.TrimSpace(p.UID), strings.TrimSpace(p.Key), strings.TrimSpace(p.Token), p.Status, strings.TrimSpace(p.Remark),
	)
	if err != nil {
		return 0, fmt.Errorf("新增连接失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (s *XMService) AdminDeleteProvider(id int) error {
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM xm_project WHERE provider_id = ? AND is_deleted = 0", id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该连接下还有 %d 个项目，请先删除或迁移项目", count)
	}
	_, err := database.DB.Exec("UPDATE xm_provider SET is_deleted = 1, updated_at = NOW() WHERE id = ?", id)
	return err
}

func (s *XMService) updateProviderSyncTime(providerID int) {
	if providerID <= 0 {
		return
	}
	database.DB.Exec("UPDATE xm_provider SET last_sync_at = NOW(), updated_at = NOW() WHERE id = ?", providerID)
}

func normalizeUpstreamProjects(result map[string]interface{}) ([]XMUpstreamProject, error) {
	code := parseXMInt(result["code"])
	if code != 1 && code != 200 && code != 0 {
		msg := strings.TrimSpace(fmt.Sprintf("%v", result["msg"]))
		if msg == "" {
			msg = "上游返回失败"
		}
		return nil, fmt.Errorf(msg)
	}

	rawList, ok := result["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("上游项目数据格式错误")
	}

	list := make([]XMUpstreamProject, 0, len(rawList))
	for _, item := range rawList {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		project := XMUpstreamProject{
			ID:          strings.TrimSpace(fmt.Sprintf("%v", row["id"])),
			Name:        strings.TrimSpace(fmt.Sprintf("%v", row["name"])),
			Description: strings.TrimSpace(fmt.Sprintf("%v", row["description"])),
			Price:       roundXMPrice(parseXMFloat(row["price"])),
			Query:       parseXMInt(row["query"]),
			Password:    parseXMInt(row["password"]),
		}
		if project.ID == "" {
			continue
		}
		list = append(list, project)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list, nil
}

func (s *XMService) fetchProjectsByConnection(conn xmConnectionConfig) ([]XMUpstreamProject, error) {
	var result map[string]interface{}
	var err error

	if conn.AuthType == 0 {
		result, err = s.providerRequest(conn, "getProjects", nil, "GET")
		if err != nil {
			return nil, err
		}
		return normalizeUpstreamProjects(result)
	}

	// token 模式的上游并没有统一协议。
	// 这里按常见 REST 风格依次尝试，尽量兼容更多来源。
	candidates := []string{
		strings.TrimRight(conn.BaseURL, "/") + "/projects",
		strings.TrimRight(conn.BaseURL, "/") + "/getProjects",
	}
	for _, reqURL := range candidates {
		result, err = s.httpRequest("GET", reqURL, nil, buildXMTokenHeaders(conn.Token, nil))
		if err != nil {
			continue
		}
		list, normalizeErr := normalizeUpstreamProjects(result)
		if normalizeErr == nil {
			return list, nil
		}
		err = normalizeErr
	}

	if err == nil {
		err = fmt.Errorf("Token 模式暂不支持自动拉取项目，请检查上游接口协议")
	}
	return nil, err
}

func (s *XMService) AdminTestProvider(providerID int) (map[string]interface{}, error) {
	conn, err := s.getProviderConnection(providerID)
	if err != nil {
		return nil, err
	}
	projects, err := s.fetchProjectsByConnection(conn)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":       "连接成功",
		"project_count": len(projects),
	}, nil
}

func (s *XMService) AdminFetchProviderProjects(providerID int) ([]XMUpstreamProject, error) {
	conn, err := s.getProviderConnection(providerID)
	if err != nil {
		return nil, err
	}
	projects, err := s.fetchProjectsByConnection(conn)
	if err != nil {
		return nil, err
	}
	s.updateProviderSyncTime(providerID)
	return projects, nil
}

func projectIDSelected(projectID string, selected map[string]struct{}) bool {
	if len(selected) == 0 {
		return true
	}
	_, ok := selected[projectID]
	return ok
}

func selectedProjectSet(ids []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			result[id] = struct{}{}
		}
	}
	return result
}

func (s *XMService) AdminImportProviderProjects(req XMImportProjectsRequest) (map[string]interface{}, error) {
	conn, err := s.getProviderConnection(req.ProviderID)
	if err != nil {
		return nil, err
	}
	projects, err := s.fetchProjectsByConnection(conn)
	if err != nil {
		return nil, err
	}

	selected := selectedProjectSet(req.ProjectIDs)
	summary := xmImportSummary{}

	for _, item := range projects {
		if !projectIDSelected(item.ID, selected) {
			continue
		}
		summary.Total++

		localPrice := localPriceByRule(item.Price, req.PriceMultiplier, req.PriceAddition)

		var existingID int
		var existingLocalPrice float64
		err := database.DB.QueryRow(
			"SELECT id, COALESCE(local_price, 0) FROM xm_project WHERE provider_id = ? AND p_id = ? AND is_deleted = 0 LIMIT 1",
			req.ProviderID, item.ID,
		).Scan(&existingID, &existingLocalPrice)

		if err == nil && existingID > 0 {
			if req.OverwriteLocalPrice {
				_, err = database.DB.Exec(
					`UPDATE xm_project
					 SET name=?, description=?, upstream_price=?, local_price=?, price=?, `+"`query`"+`=?, password=?, provider_id=?, updated_at=NOW()
					 WHERE id=?`,
					item.Name, item.Description, item.Price, localPrice, localPrice, item.Query, item.Password, req.ProviderID, existingID,
				)
			} else {
				_, err = database.DB.Exec(
					`UPDATE xm_project
					 SET name=?, description=?, upstream_price=?, `+"`query`"+`=?, password=?, provider_id=?, updated_at=NOW()
					 WHERE id=?`,
					item.Name, item.Description, item.Price, item.Query, item.Password, req.ProviderID, existingID,
				)
			}
			if err != nil {
				return nil, fmt.Errorf("更新项目失败: %v", err)
			}
			summary.Updated++
			continue
		}

		_, err = database.DB.Exec(
			`INSERT INTO xm_project
			 (name, description, price, upstream_price, local_price, `+"`query`"+`, password, p_id, status, provider_id, sort_order, sync_mode, is_deleted, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, 0, 1, 0, NOW(), NOW())`,
			item.Name,
			item.Description,
			localPrice,
			item.Price,
			localPrice,
			item.Query,
			item.Password,
			item.ID,
			req.ProviderID,
		)
		if err != nil {
			return nil, fmt.Errorf("导入项目失败: %v", err)
		}
		summary.Created++
	}

	s.updateProviderSyncTime(req.ProviderID)
	return map[string]interface{}{
		"summary": summary,
	}, nil
}

func (s *XMService) AdminSyncProviderProjects(req XMSyncProjectsRequest) (map[string]interface{}, error) {
	conn, err := s.getProviderConnection(req.ProviderID)
	if err != nil {
		return nil, err
	}
	projects, err := s.fetchProjectsByConnection(conn)
	if err != nil {
		return nil, err
	}

	upstreamMap := map[string]XMUpstreamProject{}
	for _, item := range projects {
		upstreamMap[item.ID] = item
	}

	rows, err := database.DB.Query("SELECT id, p_id FROM xm_project WHERE provider_id = ? AND is_deleted = 0", req.ProviderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updated := 0
	skipped := 0
	for rows.Next() {
		var id int
		var pID string
		if err := rows.Scan(&id, &pID); err != nil {
			continue
		}
		upstream, ok := upstreamMap[pID]
		if !ok {
			skipped++
			continue
		}

		updateParts := []string{"updated_at = NOW()"}
		updateArgs := make([]interface{}, 0)

		if req.SyncName {
			updateParts = append(updateParts, "name = ?")
			updateArgs = append(updateArgs, upstream.Name)
		}
		if req.SyncDescription {
			updateParts = append(updateParts, "description = ?")
			updateArgs = append(updateArgs, upstream.Description)
		}
		if req.SyncUpstreamPrice {
			updateParts = append(updateParts, "upstream_price = ?")
			updateArgs = append(updateArgs, upstream.Price)
		}
		if req.SyncQuery {
			updateParts = append(updateParts, "`query` = ?")
			updateArgs = append(updateArgs, upstream.Query)
		}
		if req.SyncPassword {
			updateParts = append(updateParts, "password = ?")
			updateArgs = append(updateArgs, upstream.Password)
		}
		if req.OverwriteLocalPrice {
			localPrice := localPriceByRule(upstream.Price, req.PriceMultiplier, req.PriceAddition)
			updateParts = append(updateParts, "local_price = ?", "price = ?")
			updateArgs = append(updateArgs, localPrice, localPrice)
		}

		updateArgs = append(updateArgs, id)
		sql := "UPDATE xm_project SET " + strings.Join(updateParts, ", ") + " WHERE id = ?"
		if _, err := database.DB.Exec(sql, updateArgs...); err != nil {
			return nil, fmt.Errorf("同步项目失败: %v", err)
		}
		updated++
	}

	s.updateProviderSyncTime(req.ProviderID)
	return map[string]interface{}{
		"summary": map[string]int{
			"updated": updated,
			"skipped": skipped,
			"total":   len(projects),
		},
	}, nil
}

func (s *XMService) migrateLegacyXMProviders() {
	rows, err := database.DB.Query(`
		SELECT DISTINCT COALESCE(url, ''), COALESCE(type, ''), COALESCE(uid, ''), COALESCE(` + "`key`" + `, ''), COALESCE(token, '')
		FROM xm_project
		WHERE is_deleted = 0
		  AND provider_id = 0
		  AND COALESCE(url, '') <> ''`)
	if err != nil {
		return
	}
	defer rows.Close()

	type legacyConn struct {
		URL      string
		Type     string
		UID      string
		Key      string
		Token    string
		Provider int
	}

	connections := make([]legacyConn, 0)
	for rows.Next() {
		var item legacyConn
		if err := rows.Scan(&item.URL, &item.Type, &item.UID, &item.Key, &item.Token); err != nil {
			continue
		}
		item.Type = strings.TrimSpace(item.Type)
		if item.URL == "" {
			continue
		}
		connections = append(connections, item)
	}

	for _, item := range connections {
		authType := parseXMInt(item.Type)
		var providerID int
		err := database.DB.QueryRow(
			"SELECT id FROM xm_provider WHERE is_deleted = 0 AND base_url = ? AND auth_type = ? AND uid = ? AND `key` = ? AND token = ? LIMIT 1",
			item.URL, authType, item.UID, item.Key, item.Token,
		).Scan(&providerID)
		if err != nil || providerID <= 0 {
			name := fmt.Sprintf("XM-%s", strings.TrimPrefix(strings.TrimPrefix(item.URL, "https://"), "http://"))
			result, execErr := database.DB.Exec(
				`INSERT INTO xm_provider (name, base_url, auth_type, uid, `+"`key`"+`, token, status, remark, is_deleted, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, 0, 'legacy auto migrated', 0, NOW(), NOW())`,
				name, item.URL, authType, item.UID, item.Key, item.Token,
			)
			if execErr != nil {
				continue
			}
			newID, _ := result.LastInsertId()
			providerID = int(newID)
		}
		if providerID > 0 {
			database.DB.Exec(
				"UPDATE xm_project SET provider_id = ? WHERE provider_id = 0 AND COALESCE(url, '') = ? AND COALESCE(type, '') = ? AND COALESCE(uid, '') = ? AND COALESCE(`key`, '') = ? AND COALESCE(token, '') = ?",
				providerID, item.URL, item.Type, item.UID, item.Key, item.Token,
			)
		}
	}
}
