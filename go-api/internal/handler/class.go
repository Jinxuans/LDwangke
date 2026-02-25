package handler

import (
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var classService = service.NewClassService()

func ClassList(c *gin.Context) {
	var req model.ClassListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = model.ClassListRequest{}
	}

	uid := c.GetInt("uid")
	list, err := classService.List(uid, req)
	if err != nil {
		response.ServerError(c, "查询课程失败")
		return
	}

	response.Success(c, list)
}

func ClassSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "请输入搜索关键词")
		return
	}

	uid := c.GetInt("uid")
	req := model.ClassListRequest{Search: keyword}
	list, err := classService.List(uid, req)
	if err != nil {
		response.ServerError(c, "搜索课程失败")
		return
	}

	response.Success(c, list)
}

var supplierService = service.NewSupplierService()

func ClassQueryCourse(c *gin.Context) {
	var req model.CourseQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写课程ID和下单信息")
		return
	}

	result, err := supplierService.QueryCourse(req.CID, req.UserInfo)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, result)
}

func ClassCategorySwitches(c *gin.Context) {
	cid := c.Query("cid")
	if cid == "" {
		response.BadRequest(c, "缺少cid参数")
		return
	}
	cidInt := 0
	for _, ch := range cid {
		if ch >= '0' && ch <= '9' {
			cidInt = cidInt*10 + int(ch-'0')
		}
	}
	log, ticket, changepass, allowpause, supplierReport, supplierReportHID, _ := adminService.CategorySwitchesByCID(cidInt)
	response.Success(c, gin.H{
		"log": log, "ticket": ticket, "changepass": changepass, "allowpause": allowpause, "supplier_report": supplierReport, "supplier_report_hid": supplierReportHID,
	})
}

func ClassCategories(c *gin.Context) {
	categories, err := classService.Categories()
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}

	response.Success(c, categories)
}
