package model

// ExtMenu 扩展菜单
type ExtMenu struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	URL       string `json:"url"` // 页面路径或完整 URL
	SortOrder int    `json:"sort_order"`
	Visible   int    `json:"visible"` // 1=显示 0=隐藏
	Scope     string `json:"scope"`   // frontend / backend
	CreatedAt string `json:"created_at"`
}

// ExtMenuSaveRequest 保存扩展菜单
type ExtMenuSaveRequest struct {
	ID        int    `json:"id"`
	Title     string `json:"title" binding:"required"`
	Icon      string `json:"icon"`
	URL       string `json:"url" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Visible   int    `json:"visible"`
	Scope     string `json:"scope"`
}
