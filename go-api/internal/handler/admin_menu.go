package handler

import (
	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// MenuEnsureTable 自动建表
func MenuEnsureTable() {
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS menu_config (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		menu_key   VARCHAR(100) NOT NULL UNIQUE,
		parent_key VARCHAR(100) DEFAULT '',
		title      VARCHAR(100) DEFAULT '',
		icon       VARCHAR(200) DEFAULT '',
		sort_order INT DEFAULT 0,
		visible    TINYINT DEFAULT 1,
		scope      VARCHAR(20) DEFAULT 'frontend',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
}

// AdminMenuList 获取全部菜单配置
func AdminMenuList(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, menu_key, COALESCE(parent_key,''), COALESCE(title,''), COALESCE(icon,''), sort_order, visible, COALESCE(scope,'frontend') FROM menu_config ORDER BY scope, sort_order ASC",
	)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id, sortOrder, visible int
		var menuKey, parentKey, title, icon, scope string
		rows.Scan(&id, &menuKey, &parentKey, &title, &icon, &sortOrder, &visible, &scope)
		list = append(list, gin.H{
			"id":         id,
			"menu_key":   menuKey,
			"parent_key": parentKey,
			"title":      title,
			"icon":       icon,
			"sort_order": sortOrder,
			"visible":    visible,
			"scope":      scope,
		})
	}
	if list == nil {
		list = []gin.H{}
	}
	response.Success(c, list)
}

// AdminMenuSave 批量保存菜单配置（UPSERT）
func AdminMenuSave(c *gin.Context) {
	var req struct {
		Items []struct {
			MenuKey   string `json:"menu_key"`
			ParentKey string `json:"parent_key"`
			Title     string `json:"title"`
			Icon      string `json:"icon"`
			SortOrder int    `json:"sort_order"`
			Visible   int    `json:"visible"`
			Scope     string `json:"scope"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	for _, item := range req.Items {
		if item.Scope == "" {
			item.Scope = "frontend"
		}
		_, err := database.DB.Exec(
			`INSERT INTO menu_config (menu_key, parent_key, title, icon, sort_order, visible, scope)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE parent_key=VALUES(parent_key), title=VALUES(title), icon=VALUES(icon),
			   sort_order=VALUES(sort_order), visible=VALUES(visible), scope=VALUES(scope)`,
			item.MenuKey, item.ParentKey, item.Title, item.Icon, item.SortOrder, item.Visible, item.Scope,
		)
		if err != nil {
			response.ServerError(c, "保存菜单配置失败: "+err.Error())
			return
		}
	}
	response.SuccessMsg(c, "菜单配置已保存")
}
