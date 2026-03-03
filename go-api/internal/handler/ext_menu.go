package handler

import (
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var extMenuService = service.NewExtMenuService()

// AdminExtMenuList 获取所有扩展菜单（管理员）
func AdminExtMenuList(c *gin.Context) {
	list, err := extMenuService.List()
	if err != nil {
		response.ServerError(c, "查询扩展菜单失败")
		return
	}
	response.Success(c, list)
}

// AdminExtMenuSave 保存扩展菜单（管理员）
func AdminExtMenuSave(c *gin.Context) {
	var req model.ExtMenuSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写标题和URL")
		return
	}
	if err := extMenuService.Save(req); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// AdminExtMenuDelete 删除扩展菜单（管理员）
func AdminExtMenuDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效ID")
		return
	}
	if err := extMenuService.Delete(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// AdminExtMenuReorder 批量更新扩展菜单排序
func AdminExtMenuReorder(c *gin.Context) {
	var req struct {
		Items []struct {
			ID        int `json:"id"`
			SortOrder int `json:"sort_order"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	for _, item := range req.Items {
		database.DB.Exec("UPDATE qingka_ext_menu SET sort_order=? WHERE id=?", item.SortOrder, item.ID)
	}
	response.SuccessMsg(c, "排序已更新")
}

// ExtMenuPublicList 获取可见的扩展菜单（公开，用于前端动态菜单）
func ExtMenuPublicList(c *gin.Context) {
	list, err := extMenuService.List()
	if err != nil {
		response.Success(c, []model.ExtMenu{})
		return
	}
	// 只返回可见的
	var visible []model.ExtMenu
	for _, m := range list {
		if m.Visible == 1 {
			visible = append(visible, m)
		}
	}
	if visible == nil {
		visible = []model.ExtMenu{}
	}
	response.Success(c, visible)
}
