package handler

import (
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var openAPIService = service.NewOpenAPIService()

// OpenAPIGetClass 外部API: 获取课程列表
// PHP: act=getclass / act=sxgetclass
func OpenAPIGetClass(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	list, err := openAPIService.GetClassList(uid, money)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "data": list})
}

// OpenAPIQuery 外部API: 查课
// PHP: act=get / act=sxgetclass + userinfo
func OpenAPIQuery(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	cidStr := c.Query("cid")
	if cidStr == "" {
		cidStr = c.PostForm("cid")
	}
	cid, _ := strconv.Atoi(cidStr)
	if cid <= 0 {
		response.BusinessError(c, 0, "缺少 cid 参数")
		return
	}

	userinfo := c.Query("userinfo")
	if userinfo == "" {
		userinfo = c.PostForm("userinfo")
	}
	if userinfo == "" {
		response.BusinessError(c, 0, "缺少 userinfo 参数")
		return
	}

	result, err := openAPIService.QueryCourse(uid, money, cid, userinfo)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

// OpenAPIAddOrder 外部API: 下单
// PHP: act=sxadd
func OpenAPIAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	cidStr := c.Query("cid")
	if cidStr == "" {
		cidStr = c.PostForm("cid")
	}
	cid, _ := strconv.Atoi(cidStr)
	if cid <= 0 {
		response.BusinessError(c, 0, "缺少 cid 参数")
		return
	}

	userinfo := c.Query("userinfo")
	if userinfo == "" {
		userinfo = c.PostForm("userinfo")
	}
	if userinfo == "" {
		response.BusinessError(c, 0, "缺少 userinfo 参数")
		return
	}

	// 构造订单请求
	req := model.OrderAddRequest{
		CID: cid,
		Data: []model.OrderAddItem{
			{UserInfo: userinfo},
		},
	}

	result, err := openAPIService.AddOrder(uid, money, req)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

// OpenAPIOrderList 外部API: 订单列表
// PHP: act=user_orderlist
func OpenAPIOrderList(c *gin.Context) {
	uid := c.GetInt("uid")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := openAPIService.OrderList(uid, page, limit, status)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// OpenAPIBalance 外部API: 查询余额
func OpenAPIBalance(c *gin.Context) {
	money := c.GetFloat64("money")
	response.Success(c, gin.H{"money": money})
}
