package model

type DynamicModule struct {
	ID          int    `json:"id"`
	AppID       string `json:"app_id"`
	Type        string `json:"type"` // sport/intern/paper
	Name        string `json:"name"`
	Description string `json:"description"` // 模块描述
	Price       string `json:"price"`       // 展示价格(如 0.5元/次)
	Icon        string `json:"icon"`
	ApiBase     string `json:"api_base"`
	ViewURL     string `json:"view_url"` // PHP 前端单页 URL 路径
	Status      int    `json:"status"`   // 0=禁用 1=启用
	Sort        int    `json:"sort"`
	Config      string `json:"config"` // JSON 配置（表单字段等）
}

type ModuleSaveRequest struct {
	ID          int    `json:"id"`
	AppID       string `json:"app_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Icon        string `json:"icon"`
	ApiBase     string `json:"api_base"`
	ViewURL     string `json:"view_url"`
	Status      int    `json:"status"`
	Sort        int    `json:"sort"`
	Config      string `json:"config"`
}
