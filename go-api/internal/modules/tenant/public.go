package tenant

import (
	"strconv"

	"go-api/internal/model"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// ===== B端后台（登录用户，需有店铺）=====

// GET /api/v1/tenant/mall-open-price
func TenantMallOpenPrice(c *gin.Context) {
	price := tenantService.GetMallOpenPrice()
	response.Success(c, gin.H{"price": price})
}

// POST /api/v1/tenant/mall-open
func TenantMallOpen(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		ShopName string `json:"shop_name"`
	}
	_ = c.ShouldBindJSON(&req)
	tid, err := tenantService.OpenMall(uid, req.ShopName)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, gin.H{"tid": tid})
}

func TenantShopGet(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	response.Success(c, gin.H{
		"tid":                t.TID,
		"uid":                t.UID,
		"shop_name":          t.ShopName,
		"shop_logo":          t.ShopLogo,
		"shop_desc":          t.ShopDesc,
		"domain":             t.Domain,
		"pay_config":         t.PayConfig,
		"mall_config":        t.MallConfig,
		"status":             t.Status,
		"addtime":            t.AddTime,
		"mall_domain_suffix": normalizeTenantDomain(getAdminConfigValue("mall_domain_suffix")),
	})
}

func TenantShopSave(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.TenantSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 检查是否已有店铺
	t, _ := lookupTenantByUID(uid)
	if t == nil {
		// 新建
		req.UID = uid
		tid, err := CreateTenant(uid, &req)
		if err != nil {
			response.BusinessError(c, 1001, err.Error())
			return
		}
		response.Success(c, gin.H{"tid": tid})
		return
	}
	if err := updateTenantShop(uid, &req); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TenantPayConfigSave(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.TenantPayConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := updateTenantPayConfig(uid, req.PayConfig); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TenantMallConfigSave(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.TenantMallConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	normalized, err := normalizeTenantMallConfig(req.MallConfig)
	if err != nil {
		response.BusinessError(c, 1009, err.Error())
		return
	}
	if err := updateTenantMallConfig(uid, normalized); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TenantMallCategoryList(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	list, err := listTenantMallCategories(t.TID)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func TenantMallCategorySave(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	var req model.TenantMallCategorySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := saveTenantMallCategory(t.TID, &req); err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TenantMallCategoryDelete(c *gin.Context) {
	uid := c.GetInt("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	if err := deleteTenantMallCategory(t.TID, id); err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func TenantMallCategoryUpdateSort(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	var body struct {
		Items []struct {
			ID   int `json:"id" binding:"required"`
			Sort int `json:"sort" binding:"required"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items := make([]struct{ ID, Sort int }, 0, len(body.Items))
	for _, item := range body.Items {
		items = append(items, struct{ ID, Sort int }{ID: item.ID, Sort: item.Sort})
	}
	if err := updateTenantMallCategorySort(t.TID, items); err != nil {
		response.ServerError(c, "排序更新失败")
		return
	}
	response.SuccessMsg(c, "排序更新成功")
}

func TenantProductList(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	list, err := listTenantProducts(t.TID)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func TenantProductSave(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	var req model.TenantProductSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := saveTenantProduct(t.TID, &req); err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TenantProductDelete(c *gin.Context) {
	uid := c.GetInt("uid")
	cid, _ := strconv.Atoi(c.Param("cid"))
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	if err := deleteTenantProduct(t.TID, cid); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func TenantOrderStats(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	stats, err := getTenantOrderStats(t.TID)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, stats)
}

// ===== C端用户管理 =====

// GET /api/v1/tenant/cusers
func TenantCUserList(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := listTenantCUsers(t.TID, page, limit)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// POST /api/v1/tenant/cuser/save
func TenantCUserSave(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	var req model.CUserSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := saveTenantCUser(t.TID, &req); err != nil {
		response.BusinessError(c, 1007, err.Error())
		return
	}
	response.Success(c, nil)
}

// DELETE /api/v1/tenant/cuser/:id
func TenantCUserDelete(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := deleteTenantCUser(t.TID, id); err != nil {
		response.BusinessError(c, 1008, err.Error())
		return
	}
	response.Success(c, nil)
}

// GET /api/v1/tenant/mall-orders
func TenantMallOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := listTenantMallOrders(t.TID, page, limit)
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func TenantMallLinkedOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "支付订单ID无效")
		return
	}
	list, err := listTenantMallLinkedOrders(t.TID, id)
	if err != nil {
		response.ServerError(c, "查询关联订单失败")
		return
	}
	response.Success(c, list)
}

func TenantCUserWithdrawRequests(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	var req model.AdminCUserWithdrawListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := tenantService.TenantCUserWithdrawRequests(t.TID, req)
	if err != nil {
		response.ServerError(c, "查询会员提现记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func TenantCUserWithdrawReview(c *gin.Context) {
	uid := c.GetInt("uid")
	t, err := lookupTenantByUID(uid)
	if err != nil {
		response.BusinessError(c, 1002, "未开通店铺")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "提现申请ID无效")
		return
	}
	var req model.WithdrawReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "审核参数错误")
		return
	}
	if err := tenantService.ReviewTenantCUserWithdrawRequest(uid, t.TID, id, req.Status, req.Remark); err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.SuccessMsg(c, "审核完成")
}
