package admin

import (
	"strconv"

	"go-api/internal/model"
	suppliermodule "go-api/internal/modules/supplier"
	wmodule "go-api/internal/modules/w"
	xmmodule "go-api/internal/modules/xm"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var adminSupplierToolService = suppliermodule.SharedService()

type CloneRequest struct {
	HID             int               `json:"hid"`
	PriceRate       float64           `json:"price_rate"`
	CloneCategory   bool              `json:"clone_category"`
	SkipCategories  []string          `json:"skip_categories"`
	NameReplace     map[string]string `json:"name_replace"`
	SecretPriceRate float64           `json:"secret_price_rate"`
}

func registerSupplierRoutes(admin *gin.RouterGroup) {
	admin.GET("/suppliers", AdminSupplierList)
	admin.POST("/supplier/save", AdminSupplierSave)
	admin.POST("/supplier/delete", AdminSupplierDelete)
	admin.DELETE("/supplier/:hid", AdminSupplierDelete)
	admin.GET("/supplier/balance", suppliermodule.AdminSupplierBalance)
	admin.GET("/supplier/import", suppliermodule.AdminSupplierImport)
	admin.GET("/supplier/sync-status", suppliermodule.AdminSupplierSyncStatus)
	admin.GET("/supplier/products", suppliermodule.AdminSupplierProducts)

	admin.POST("/clone/execute", AdminCloneExecute)
	admin.POST("/clone/update-prices", AdminCloneUpdatePrices)
	admin.POST("/clone/auto-sync", AdminCloneAutoSync)
	admin.GET("/platform-names", suppliermodule.AdminPlatformNames)

	admin.GET("/xm-project/list", XMAdminListProjects)
	admin.POST("/xm-project/save", XMAdminSaveProject)
	admin.DELETE("/xm-project/delete", XMAdminDeleteProject)
	admin.GET("/w-app/list", WAdminListApps)
	admin.POST("/w-app/save", WAdminSaveApp)
	admin.DELETE("/w-app/delete", WAdminDeleteApp)

	admin.POST("/xuemei/shouhou", XueMeiShouHou)
	admin.GET("/xuemei/getcity", XueMeiGetCity)
	admin.POST("/xuemei/getcity", XueMeiGetCity)
	admin.POST("/xuemei/editip", XueMeiEditIP)
	admin.POST("/xuemei/youxian", XueMeiYouXian)
	admin.GET("/xuemei/getname", XueMeiGetName)
	admin.POST("/xuemei/getname", XueMeiGetName)
	admin.POST("/xuemei/editname", XueMeiEditName)
	admin.GET("/xuemei/zhs-log", XueMeiChaZhsLog)
	admin.POST("/xuemei/zhs-log", XueMeiChaZhsLog)
}

func AdminSupplierList(c *gin.Context) {
	list, err := adminSupplierToolService.ListSuppliers()
	if err != nil {
		response.ServerError(c, "查询货源失败")
		return
	}
	response.Success(c, list)
}

func AdminSupplierSave(c *gin.Context) {
	var sup model.Supplier
	if err := c.ShouldBindJSON(&sup); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminSupplierToolService.SaveSupplier(sup); err != nil {
		response.ServerError(c, "保存货源失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminSupplierDelete(c *gin.Context) {
	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		response.BadRequest(c, "无效的供应商ID")
		return
	}
	if err := adminSupplierToolService.DeleteSupplier(hid); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminCloneExecute(c *gin.Context) {
	req, ok := bindCloneRequest(c)
	if !ok {
		return
	}

	result, err := adminSyncExecute(req.HID, buildCloneSyncConfig(req, cloneSyncModeCloneOnly))
	if err != nil {
		response.ServerError(c, "克隆失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "克隆完成", "result": result})
}

func AdminCloneUpdatePrices(c *gin.Context) {
	req, ok := bindCloneRequest(c)
	if !ok {
		return
	}

	result, err := adminSyncExecute(req.HID, buildCloneSyncConfig(req, cloneSyncModePriceOnly))
	if err != nil {
		response.ServerError(c, "更新价格失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "价格更新完成", "result": result})
}

func AdminCloneAutoSync(c *gin.Context) {
	req, ok := bindCloneRequest(c)
	if !ok {
		return
	}

	result, err := adminSyncExecute(req.HID, buildCloneSyncConfig(req, cloneSyncModeFull))
	if err != nil {
		response.ServerError(c, "同步失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "同步完成", "result": result})
}

func XMAdminListProjects(c *gin.Context) {
	list, err := xmmodule.XM().AdminListProjects()
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func XMAdminSaveProject(c *gin.Context) {
	var project xmmodule.XMProjectAdmin
	if err := c.ShouldBindJSON(&project); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}
	id, err := xmmodule.XM().AdminSaveProject(project)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, map[string]int{"id": id})
}

func XMAdminDeleteProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.BadRequest(c, "缺少项目ID")
		return
	}
	if err := xmmodule.XM().AdminDeleteProject(id); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func WAdminListApps(c *gin.Context) {
	list, err := wmodule.W().AdminListApps()
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func WAdminSaveApp(c *gin.Context) {
	var app wmodule.WApp
	if err := c.ShouldBindJSON(&app); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	id, err := wmodule.W().AdminSaveApp(app)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, map[string]int64{"id": id})
}

func WAdminDeleteApp(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		response.BadRequest(c, "缺少ID")
		return
	}
	if err := wmodule.W().AdminDeleteApp(id); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func XueMeiShouHou(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		FanKui string `json:"fankui" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、fankui")
		return
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	code, msg, err := suppliermodule.XueMeiShouHou(sup, req.OID, req.FanKui)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

func XueMeiGetCity(c *gin.Context) {
	var req struct {
		HID int `json:"hid" form:"hid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid")
			return
		}
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	data, err := suppliermodule.XueMeiGetCity(sup)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

func XueMeiEditIP(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		NodeID string `json:"node_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、node_id")
		return
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	code, msg, err := suppliermodule.XueMeiEditIP(sup, req.OID, req.NodeID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

func XueMeiYouXian(c *gin.Context) {
	var req struct {
		HID int    `json:"hid" binding:"required"`
		OID string `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid")
		return
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	code, msg, err := suppliermodule.XueMeiYouXian(sup, req.OID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

func XueMeiGetName(c *gin.Context) {
	var req struct {
		HID     int    `json:"hid" form:"hid" binding:"required"`
		OrderID string `json:"order_id" form:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid、order_id")
			return
		}
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	data, err := suppliermodule.XueMeiGetName(sup, req.OrderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

func XueMeiEditName(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		NameID string `json:"name_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、name_id")
		return
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	code, msg, err := suppliermodule.XueMeiEditName(sup, req.OID, req.NameID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

func XueMeiChaZhsLog(c *gin.Context) {
	var req struct {
		HID     int    `json:"hid" form:"hid" binding:"required"`
		OrderID string `json:"order_id" form:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid、order_id")
			return
		}
	}

	sup, ok := getXueMeiSupplier(c, req.HID)
	if !ok {
		return
	}

	data, err := suppliermodule.XueMeiChaZhsLog(sup, req.OrderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

func getXueMeiSupplier(c *gin.Context, hid int) (*model.SupplierFull, bool) {
	sup, err := adminSupplierToolService.GetSupplierByHID(hid)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return nil, false
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return nil, false
	}
	return sup, true
}

type cloneSyncMode int

const (
	cloneSyncModeCloneOnly cloneSyncMode = iota
	cloneSyncModePriceOnly
	cloneSyncModeFull
)

func bindCloneRequest(c *gin.Context) (CloneRequest, bool) {
	var req CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return CloneRequest{}, false
	}
	if req.HID <= 0 {
		response.BadRequest(c, "货源ID不能为空")
		return CloneRequest{}, false
	}
	if req.PriceRate <= 0 {
		req.PriceRate = 5
	}
	return req, true
}

func buildCloneSyncConfig(req CloneRequest, mode cloneSyncMode) *SyncConfig {
	cfg := &SyncConfig{
		SupplierIDs: strconv.Itoa(req.HID),
		PriceRates: map[string]float64{
			strconv.Itoa(req.HID): req.PriceRate,
		},
		SkipCategories:  req.SkipCategories,
		NameReplace:     req.NameReplace,
		SecretPriceRate: req.SecretPriceRate,
	}

	switch mode {
	case cloneSyncModeCloneOnly:
		cfg.CloneEnabled = true
		cfg.CloneCategory = req.CloneCategory
	case cloneSyncModePriceOnly:
		cfg.SyncPrice = true
	case cloneSyncModeFull:
		cfg.CloneEnabled = true
		cfg.CloneCategory = true
		cfg.SyncPrice = true
		cfg.SyncStatus = true
		cfg.SyncContent = true
		cfg.SyncName = false
	}

	return cfg
}
