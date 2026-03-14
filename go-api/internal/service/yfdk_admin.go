package service

import (
	"fmt"
	"log"
	"strings"

	"go-api/internal/database"
)

func (s *YFDKService) GetAdminProjects() ([]YFDKProject, error) {
	rows, err := database.DB.Query("SELECT id, cid, name, content, cost_price, sell_price, enabled, sort, create_time, update_time FROM qingka_wangke_yfdk_projects ORDER BY sort ASC, id ASC")
	if err != nil {
		log.Printf("[YFDK] 查询项目列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var projects []YFDKProject
	for rows.Next() {
		var p YFDKProject
		var createTime, updateTime []uint8
		err := rows.Scan(&p.ID, &p.CID, &p.Name, &p.Content, &p.CostPrice, &p.SellPrice, &p.Enabled, &p.Sort, &createTime, &updateTime)
		if err != nil {
			log.Printf("[YFDK] 扫描项目行失败: %v", err)
			continue
		}
		p.CreateTime = string(createTime)
		p.UpdateTime = string(updateTime)
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []YFDKProject{}
	}
	return projects, nil
}

func (s *YFDKService) SyncProjectsFromUpstream() (int, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return 0, fmt.Errorf("YF打卡未配置")
	}

	result, err := s.upstreamRequest("GET", strings.TrimRight(cfg.BaseURL, "/")+"/projects", nil, cfg.Token)
	if err != nil {
		return 0, err
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		return 0, fmt.Errorf("获取上游项目列表失败")
	}

	data, _ := result["data"].(map[string]interface{})
	projects, _ := data["projects"].([]interface{})
	if projects == nil {
		return 0, nil
	}

	count := 0
	for _, p := range projects {
		proj, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		cid := fmt.Sprintf("%v", proj["cid"])
		name := fmt.Sprintf("%v", proj["name"])
		content := fmt.Sprintf("%v", proj["content"])
		costPrice := 0.0
		if cp, ok := proj["cost_price"].(float64); ok {
			costPrice = cp
		}
		sellPrice := 0.10
		if sp, ok := proj["sell_price"].(float64); ok {
			sellPrice = sp
		}

		var existCount int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_yfdk_projects WHERE cid = ?", cid).Scan(&existCount)
		if existCount > 0 {
			continue
		}

		_, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_yfdk_projects (cid, name, content, cost_price, sell_price, enabled, sort) VALUES (?, ?, ?, ?, ?, 1, 10)",
			cid, name, content, costPrice, sellPrice,
		)
		if err == nil {
			count++
		}
	}
	return count, nil
}

func (s *YFDKService) UpdateProject(id int, sellPrice float64, enabled int, sort int, content string) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_yfdk_projects SET sell_price = ?, enabled = ?, sort = ?, content = ? WHERE id = ?",
		sellPrice, enabled, sort, content, id,
	)
	return err
}

func (s *YFDKService) DeleteProject(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_yfdk_projects WHERE id = ?", id)
	return err
}
