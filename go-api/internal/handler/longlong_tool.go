package handler

import (
	"encoding/json"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// LonglongToolConfig 龙龙一键对接工具配置
type LonglongToolConfig struct {
	LongHost      string `json:"long_host"`
	AccessKey     string `json:"access_key"`
	MysqlHost     string `json:"mysql_host"`
	MysqlPort     string `json:"mysql_port"`
	MysqlUser     string `json:"mysql_user"`
	MysqlPassword string `json:"mysql_password"`
	MysqlDatabase string `json:"mysql_database"`
	ClassTable    string `json:"class_table"`
	OrderTable    string `json:"order_table"`
	Docking       string `json:"docking"`
	Rate          string `json:"rate"`
	NamePrefix    string `json:"name_prefix"`
	Category      string `json:"category"`
	CoverPrice    bool   `json:"cover_price"`
	CoverDesc     bool   `json:"cover_desc"`
	CoverName     bool   `json:"cover_name"`
	Sort          string `json:"sort"`
	CronValue     string `json:"cron_value"`
	CronUnit      string `json:"cron_unit"`
}

// LonglongToolGetConfig 获取龙龙对接配置
func LonglongToolGetConfig(c *gin.Context) {
	var raw string
	err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = 'longlong_tool_config'").Scan(&raw)
	if err != nil || raw == "" {
		// 返回默认配置
		response.Success(c, LonglongToolConfig{
			MysqlHost: "127.0.0.1",
			MysqlPort: "3306",
			Rate:      "1.5",
			Sort:      "0",
			CronValue: "30",
			CronUnit:  "minute",
		})
		return
	}
	var cfg LonglongToolConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		response.Success(c, LonglongToolConfig{
			MysqlHost: "127.0.0.1",
			MysqlPort: "3306",
			Rate:      "1.5",
			Sort:      "0",
			CronValue: "30",
			CronUnit:  "minute",
		})
		return
	}

	if cfg.CronValue == "" {
		cfg.CronValue = "30"
	}
	if cfg.CronUnit == "" {
		cfg.CronUnit = "minute"
	}

	response.Success(c, cfg)
}

// LonglongToolSaveConfig 保存龙龙对接配置
func LonglongToolSaveConfig(c *gin.Context) {
	var cfg LonglongToolConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (`v`, `k`) VALUES ('longlong_tool_config', ?) ON DUPLICATE KEY UPDATE `k` = ?",
		string(data), string(data),
	)
	if err != nil {
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}
