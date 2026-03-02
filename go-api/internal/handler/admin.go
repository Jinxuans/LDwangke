package handler

import (
	"fmt"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/queue"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var adminService = service.NewAdminService()

// ===== 仪表盘 =====

func AdminDashboard(c *gin.Context) {
	stats, err := adminService.DashboardStats()
	if err != nil {
		response.ServerError(c, "查询统计失败")
		return
	}
	response.Success(c, stats)
}

// ===== 消费排行榜 =====

func TopConsumers(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	list := adminService.GetTopConsumers(period)
	response.Success(c, list)
}

// ===== 用户管理 =====

func AdminUserList(c *gin.Context) {
	var req model.UserListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := adminService.UserList(req)
	if err != nil {
		fmt.Printf("[UserList ERROR] %v\n", err)
		response.ServerError(c, "查询用户失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminUserResetPass(c *gin.Context) {
	var body struct {
		UID     int    `json:"uid" binding:"required"`
		NewPass string `json:"new_pass"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.UserResetPassword(body.UID, body.NewPass); err != nil {
		response.ServerError(c, "重置密码失败")
		return
	}
	response.SuccessMsg(c, "密码已重置")
}

func AdminUserSetBalance(c *gin.Context) {
	var body struct {
		UID     int     `json:"uid" binding:"required"`
		Balance float64 `json:"balance"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.UserSetBalance(body.UID, body.Balance); err != nil {
		response.ServerError(c, "设置余额失败")
		return
	}
	response.SuccessMsg(c, "余额已更新")
}

func AdminUserSetGrade(c *gin.Context) {
	var body struct {
		UID      int     `json:"uid" binding:"required"`
		AddPrice float64 `json:"addprice"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.AddPrice < 0.01 {
		response.BadRequest(c, "费率不能小于0.01")
		return
	}
	if err := adminService.UserSetGrade(body.UID, body.AddPrice); err != nil {
		response.ServerError(c, "设置等级失败")
		return
	}
	response.SuccessMsg(c, "等级已更新")
}

// ===== 分类管理 =====

func AdminCategoryList(c *gin.Context) {
	list, err := adminService.CategoryListAll()
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	response.Success(c, list)
}

func AdminCategoryListPaged(c *gin.Context) {
	var req model.CategoryListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := adminService.CategoryListPaged(req)
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminCategorySave(c *gin.Context) {
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.CategorySave(cat); err != nil {
		response.ServerError(c, "保存分类失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminCategoryDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的分类ID")
		return
	}
	if err := adminService.CategoryDelete(id); err != nil {
		response.ServerError(c, "删除分类失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminCategoryQuickModify(c *gin.Context) {
	var body struct {
		Keyword    string `json:"keyword" binding:"required"`
		CategoryID int    `json:"category_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	affected, err := adminService.CategoryQuickModify(body.Keyword, body.CategoryID)
	if err != nil {
		response.ServerError(c, "修改失败")
		return
	}
	response.Success(c, gin.H{
		"affected": affected,
		"msg":      fmt.Sprintf("已将 %d 个包含「%s」的商品归入分类 %d", affected, body.Keyword, body.CategoryID),
	})
}

func AdminCategoryUpdateSort(c *gin.Context) {
	var body struct {
		Items []struct {
			ID   int `json:"id" binding:"required"`
			Sort int `json:"sort" binding:"required"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(body.Items) == 0 {
		response.BadRequest(c, "没有要更新的分类")
		return
	}
	// 直接创建符合service期望类型的切片
	var items []struct{ ID, Sort int }
	for _, item := range body.Items {
		items = append(items, struct{ ID, Sort int }{ID: item.ID, Sort: item.Sort})
	}
	if err := adminService.CategoryUpdateSort(items); err != nil {
		response.ServerError(c, "排序更新失败")
		return
	}
	response.Success(c, gin.H{"msg": "排序更新成功"})
}

// ===== 课程管理 =====

func AdminClassList(c *gin.Context) {
	cateID, _ := strconv.Atoi(c.Query("cateId"))
	keywords := c.Query("keywords")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	list, total, err := adminService.ClassList(cateID, keywords, page, limit)
	if err != nil {
		response.ServerError(c, "查询课程失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func AdminClassSave(c *gin.Context) {
	var req model.ClassEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.ClassSave(req); err != nil {
		response.ServerError(c, "保存课程失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminClassToggle(c *gin.Context) {
	var body struct {
		CID    int `json:"cid" binding:"required"`
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.ClassToggleStatus(body.CID, body.Status); err != nil {
		response.ServerError(c, "更新状态失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func AdminClassBatchDelete(c *gin.Context) {
	var body struct {
		CIDs []int `json:"cids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.CIDs) == 0 {
		response.BadRequest(c, "请选择要删除的课程")
		return
	}
	deleted, err := adminService.ClassBatchDelete(body.CIDs)
	if err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": deleted, "msg": fmt.Sprintf("成功删除 %d 个课程", deleted)})
}

func AdminClassBatchCategory(c *gin.Context) {
	var body struct {
		CIDs   []int  `json:"cids" binding:"required"`
		CateID string `json:"cateId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.CIDs) == 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	updated, err := adminService.ClassBatchCategory(body.CIDs, body.CateID)
	if err != nil {
		response.ServerError(c, "批量修改分类失败")
		return
	}
	response.Success(c, gin.H{"updated": updated, "msg": fmt.Sprintf("成功修改 %d 个课程的分类", updated)})
}

func AdminClassBatchPrice(c *gin.Context) {
	var body struct {
		CIDs    []int   `json:"cids" binding:"required"`
		Rate    float64 `json:"rate" binding:"required"`
		Yunsuan string  `json:"yunsuan"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.CIDs) == 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.Yunsuan == "" {
		body.Yunsuan = "*"
	}
	updated, err := adminService.ClassBatchPrice(body.CIDs, body.Rate, body.Yunsuan)
	if err != nil {
		response.ServerError(c, "批量修改价格失败")
		return
	}
	response.Success(c, gin.H{"updated": updated, "msg": fmt.Sprintf("成功修改 %d 个课程的价格", updated)})
}

// ===== 货源管理 =====

func AdminSupplierList(c *gin.Context) {
	list, err := adminService.SupplierList()
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
	if err := adminService.SupplierSave(sup); err != nil {
		response.ServerError(c, "保存货源失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ===== 对接插件 =====

var supService = service.NewSupplierService()

// AdminSupplierProducts 拉取上游供应商商品列表 (按 PHP getclassdata)
func AdminSupplierProducts(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}
	sup, err := supService.GetSupplierByHID(hid)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	classes, err := supService.GetSupplierClasses(sup)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	// 标记本地已存在的 (states=1)
	localNouns := map[string]bool{}
	rows, dbErr := database.DB.Query("SELECT noun FROM qingka_wangke_class WHERE docking = ? AND status >= 0", hid)
	if dbErr == nil {
		defer rows.Close()
		for rows.Next() {
			var noun string
			rows.Scan(&noun)
			localNouns[noun] = true
		}
	}
	type ProductItem struct {
		CID          string  `json:"cid"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Fenlei       string  `json:"fenlei"`
		Content      string  `json:"content"`
		CategoryName string  `json:"category_name"`
		States       int     `json:"states"`
		Sort         int     `json:"sort"`
	}
	var list []ProductItem
	for _, item := range classes {
		states := 0
		if localNouns[item.CID] {
			states = 1
		}
		list = append(list, ProductItem{
			CID:          item.CID,
			Name:         item.Name,
			Price:        item.Price,
			Fenlei:       item.Fenlei,
			Content:      item.Content,
			CategoryName: item.CategoryName,
			States:       states,
			Sort:         10,
		})
	}
	if list == nil {
		list = []ProductItem{}
	}
	response.Success(c, list)
}

// AdminAddClass 单个添加课程 (按 PHP upclass case)
func AdminAddClass(c *gin.Context) {
	var req struct {
		Sort      string `json:"sort"`
		Name      string `json:"name" binding:"required"`
		Price     string `json:"price" binding:"required"`
		GetNoun   string `json:"getnoun"`
		Noun      string `json:"noun"`
		Content   string `json:"content"`
		QueryPlat string `json:"queryplat"`
		Docking   string `json:"docking"`
		Yunsuan   string `json:"yunsuan"`
		Status    string `json:"status"`
		Fenlei    string `json:"fenlei"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Sort == "" {
		req.Sort = "10"
	}
	if req.Yunsuan == "" {
		req.Yunsuan = "*"
	}
	if req.Status == "" {
		req.Status = "1"
	}
	sortVal, _ := strconv.Atoi(req.Sort)
	statusVal, _ := strconv.Atoi(req.Status)

	// 检查是否已存在
	if req.Docking != "" && req.Docking != "0" && req.Noun != "" {
		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE docking=? AND noun=?", req.Docking, req.Noun).Scan(&cnt)
		if cnt > 0 {
			// 更新
			_, err := database.DB.Exec(
				"UPDATE qingka_wangke_class SET name=?, price=?, getnoun=?, content=?, queryplat=?, yunsuan=?, status=?, sort=?, fenlei=? WHERE docking=? AND noun=?",
				req.Name, req.Price, req.GetNoun, req.Content, req.QueryPlat, req.Yunsuan, statusVal, sortVal, req.Fenlei, req.Docking, req.Noun,
			)
			if err != nil {
				response.ServerError(c, "更新失败: "+err.Error())
				return
			}
			response.SuccessMsg(c, "已更新")
			return
		}
	}

	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_class (name, price, getnoun, noun, content, queryplat, docking, yunsuan, status, sort, fenlei, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,NOW())",
		req.Name, req.Price, req.GetNoun, req.Noun, req.Content, req.QueryPlat, req.Docking, req.Yunsuan, statusVal, sortVal, req.Fenlei,
	)
	if err != nil {
		response.ServerError(c, "添加失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "添加成功")
}

func AdminSupplierImport(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	pricee, _ := strconv.ParseFloat(c.Query("pricee"), 64)
	category := c.Query("category")
	name := c.Query("name")
	fd, _ := strconv.Atoi(c.Query("fd"))

	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}
	if pricee <= 0 {
		pricee = 1
	}
	if category == "" {
		category = "999999"
	}

	inserted, updated, msg, err := supService.ImportSupplierClasses(hid, pricee, category, name, fd)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"inserted": inserted,
		"updated":  updated,
		"msg":      msg,
	})
}

// 课程状态同步 (按 PHP updateStatus case)
func AdminSupplierSyncStatus(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}

	count, msg, err := supService.SyncSupplierStatus(hid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"count": count,
		"msg":   msg,
	})
}

// 批量分类上下架 (按 PHP yjflxj/yjflsj case)
func AdminCategoryBatchToggle(c *gin.Context) {
	var body struct {
		Action string `json:"action"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var targetStatus int
	var sourceStatus int
	if body.Action == "down" {
		sourceStatus = 0
		targetStatus = 0
	} else {
		sourceStatus = 1
		targetStatus = 1
	}

	rows, err := database.DB.Query("SELECT id FROM qingka_wangke_fenlei WHERE status = ?", sourceStatus)
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var fid int
		rows.Scan(&fid)
		database.DB.Exec("UPDATE qingka_wangke_class SET status = ? WHERE fenlei = ?", targetStatus, fid)
		count++
	}

	response.Success(c, gin.H{
		"count": count,
		"msg":   fmt.Sprintf("已%s%d个分类内的全部项目", map[string]string{"up": "上架", "down": "下架"}[body.Action], count),
	})
}

// ===== 商品下拉列表（密价用） =====

func AdminClassDropdown(c *gin.Context) {
	rows, err := database.DB.Query("SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(fenlei,'') FROM qingka_wangke_class WHERE status >= 0 ORDER BY sort ASC, cid ASC")
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	defer rows.Close()
	type ClassOpt struct {
		CID    int    `json:"cid"`
		Name   string `json:"name"`
		Price  string `json:"price"`
		Fenlei string `json:"fenlei"`
	}
	var list []ClassOpt
	for rows.Next() {
		var opt ClassOpt
		rows.Scan(&opt.CID, &opt.Name, &opt.Price, &opt.Fenlei)
		list = append(list, opt)
	}
	if list == nil {
		list = []ClassOpt{}
	}
	response.Success(c, list)
}

// ===== 管理员余额流水 =====

func AdminMoneyLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	uid := c.Query("uid")
	logType := c.Query("type")
	list, total, err := adminService.AdminMoneyLogList(page, limit, uid, logType)
	if err != nil {
		fmt.Printf("[MoneyLog ERROR] %v\n", err)
		response.ServerError(c, "查询流水失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

// ===== 公告管理 =====

func AdminAnnouncementList(c *gin.Context) {
	var req model.AnnouncementListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := adminService.AnnouncementList(req)
	if err != nil {
		response.ServerError(c, "查询公告失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminAnnouncementSave(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")
	var req model.AnnouncementSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写标题和内容")
		return
	}
	if err := adminService.AnnouncementSave(uid, username, req); err != nil {
		response.ServerError(c, "保存公告失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminAnnouncementDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的公告ID")
		return
	}
	if err := adminService.AnnouncementDelete(id); err != nil {
		response.ServerError(c, "删除公告失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// ===== 统计报表 =====

func AdminStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}
	stats, err := adminService.StatsReport(days)
	if err != nil {
		response.ServerError(c, "查询统计失败")
		return
	}
	response.Success(c, stats)
}

// ===== 支付配置 (paydata) =====

func AdminPayDataGet(c *gin.Context) {
	data, err := adminService.GetPayData()
	if err != nil {
		response.ServerError(c, "查询支付配置失败")
		return
	}
	response.Success(c, data)
}

func AdminPayDataSave(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.SavePayData(data); err != nil {
		response.ServerError(c, "保存支付配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ===== 系统设置 =====

// 公开站点配置（无需管理员权限）
func SiteConfigGet(c *gin.Context) {
	config, err := adminService.GetPublicConfig()
	if err != nil {
		response.ServerError(c, "查询站点配置失败")
		return
	}
	response.Success(c, config)
}

func AdminConfigGet(c *gin.Context) {
	config, err := adminService.GetConfig()
	if err != nil {
		response.ServerError(c, "查询设置失败")
		return
	}
	response.Success(c, config)
}

func AdminConfigSave(c *gin.Context) {
	var configs map[string]string
	if err := c.ShouldBindJSON(&configs); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.SaveConfig(configs); err != nil {
		response.ServerError(c, "保存设置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ===== 等级管理 =====

func AdminGradeList(c *gin.Context) {
	list, err := adminService.GradeList()
	if err != nil {
		response.ServerError(c, "查询等级列表失败")
		return
	}
	response.Success(c, list)
}

func AdminGradeSave(c *gin.Context) {
	var req model.GradeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.GradeSave(req); err != nil {
		fmt.Printf("[GradeSave ERROR] %v\n", err)
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminGradeDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := adminService.GradeDelete(id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// ===== 密价设置 =====

func AdminMiJiaList(c *gin.Context) {
	var req model.MiJiaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, uids, err := adminService.MiJiaList(req)
	if err != nil {
		response.ServerError(c, "查询密价列表失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
		"uids":       uids,
	})
}

func AdminMiJiaSave(c *gin.Context) {
	var req model.MiJiaSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.MiJiaSave(req); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminMiJiaDelete(c *gin.Context) {
	var req struct {
		Mids []int `json:"mids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := adminService.MiJiaDelete(req.Mids); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminMiJiaBatch(c *gin.Context) {
	var req model.MiJiaBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	count, err := adminService.MiJiaBatch(req)
	if err != nil {
		response.ServerError(c, "批量设置失败")
		return
	}
	response.Success(c, gin.H{"count": count, "msg": fmt.Sprintf("成功为 %d 个商品设置密价", count)})
}

// AdminPlatformNames 返回已注册平台名称映射 (按 PHP xdjk.php wkname())
func AdminPlatformNames(c *gin.Context) {
	response.Success(c, service.GetPlatformNames())
}

// AdminSupplierBalance 查询供应商余额
func AdminSupplierBalance(c *gin.Context) {
	hid, err := strconv.Atoi(c.Query("hid"))
	if hid <= 0 || err != nil {
		response.BadRequest(c, "请指定供应商hid")
		return
	}
	result, err := supService.QueryBalance(hid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// AdminSupplierDelete 删除供应商
func AdminSupplierDelete(c *gin.Context) {
	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		response.BadRequest(c, "无效的供应商ID")
		return
	}
	if err := adminService.SupplierDelete(hid); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// ===== 订单操作：暂停/改密/补单 =====

func OrderPause(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少oid参数")
		return
	}

	// 检查分类开关
	oidInt, _ := strconv.Atoi(oid)
	_, _, _, allowpause, _, _, _ := adminService.CategorySwitchesByOID(oidInt)
	if allowpause == 0 {
		response.BadRequest(c, "该分类不允许暂停操作")
		return
	}

	var order struct {
		HID int    `db:"hid"`
		YID string `db:"yid"`
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,'') FROM qingka_wangke_order WHERE oid=?", oid).Scan(&order.HID, &order.YID)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法暂停")
		return
	}

	sup, err := supService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := supService.PauseOrder(sup, order.YID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderChangePassword(c *gin.Context) {
	var body struct {
		OID    int    `json:"oid" binding:"required"`
		NewPwd string `json:"newPwd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(body.NewPwd) < 3 {
		response.BadRequest(c, "密码长度至少3位")
		return
	}

	// 检查分类开关
	_, _, changepass, _, _, _, _ := adminService.CategorySwitchesByOID(body.OID)
	if changepass == 0 {
		response.BadRequest(c, "该分类不允许修改密码")
		return
	}

	var order struct {
		HID    int    `db:"hid"`
		YID    string `db:"yid"`
		Status string `db:"status"`
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid=?", body.OID).Scan(&order.HID, &order.YID, &order.Status)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.Status == "已退款" || order.Status == "已取消" {
		response.BadRequest(c, "该订单状态不允许修改密码")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法改密")
		return
	}

	sup, err := supService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := supService.ChangePassword(sup, order.YID, body.NewPwd)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET pass=? WHERE oid=?", body.NewPwd, body.OID)
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderResubmit(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少oid参数")
		return
	}

	var order struct {
		HID    int    `db:"hid"`
		YID    string `db:"yid"`
		Status string `db:"status"`
		BSNum  int    `db:"bsnum"`
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,''), COALESCE(bsnum,0) FROM qingka_wangke_order WHERE oid=?", oid).Scan(&order.HID, &order.YID, &order.Status, &order.BSNum)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.Status == "已退款" || order.Status == "已取消" {
		response.BadRequest(c, "该订单状态不允许补单")
		return
	}
	if order.BSNum > 20 {
		response.BadRequest(c, "该订单补刷已超过20次")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法补单")
		return
	}

	sup, err := supService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := supService.ResubmitOrder(sup, order.YID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET status='补刷中', dockstatus=1, bsnum=bsnum+1 WHERE oid=?", oid)
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

// ===== PUP 扩展操作：重置分数/时长/周期 =====

func OrderPupReset(c *gin.Context) {
	var body struct {
		OID   int    `json:"oid" binding:"required"`
		Type  string `json:"type" binding:"required"` // score, duration, period
		Value int    `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误：需要 oid, type 和 value")
		return
	}

	if body.Type != "score" && body.Type != "duration" && body.Type != "period" {
		response.BadRequest(c, "不支持的重置类型")
		return
	}

	var order struct {
		HID    int
		YID    string
		Status string
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid=?", body.OID).Scan(&order.HID, &order.YID, &order.Status)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法操作")
		return
	}

	sup, err := supService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	var code int
	var msg string

	switch body.Type {
	case "score":
		code, msg, err = supService.ResetOrderScore(sup, order.YID, body.Value)
	case "duration":
		code, msg, err = supService.ResetOrderDuration(sup, order.YID, body.Value)
	case "period":
		code, msg, err = supService.ResetOrderPeriod(sup, order.YID, body.Value)
	}

	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderLogs(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少 oid 参数")
		return
	}

	// 检查分类开关
	oidInt, _ := strconv.Atoi(oid)
	logSwitch, _, _, _, _, _, _ := adminService.CategorySwitchesByOID(oidInt)
	if logSwitch == 0 {
		response.BadRequest(c, "该分类未开启日志功能")
		return
	}

	var order struct {
		HID    int
		YID    string
		User   string
		Pass   string
		KCName string
		KCID   string
	}
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcname,''), COALESCE(kcid,'') FROM qingka_wangke_order WHERE oid=?", oid,
	).Scan(&order.HID, &order.YID, &order.User, &order.Pass, &order.KCName, &order.KCID)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法查看日志")
		return
	}

	sup, err := supService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	extra := map[string]string{
		"user": order.User, "pass": order.Pass,
		"kcname": order.KCName, "kcid": order.KCID,
	}
	logs, err := supService.QueryOrderLogs(sup, order.YID, extra)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if logs == nil {
		logs = []service.OrderLogEntry{}
	}
	response.Success(c, logs)
}

// ===== 对接队列管理 =====

func AdminQueueStats(c *gin.Context) {
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "队列未初始化")
		return
	}
	response.Success(c, queue.GlobalDockQueue.Stats())
}

func AdminQueueSetConcurrency(c *gin.Context) {
	var body struct {
		MaxWorkers int `json:"max_workers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "队列未初始化")
		return
	}
	queue.GlobalDockQueue.SetMaxWorkers(body.MaxWorkers)
	response.Success(c, queue.GlobalDockQueue.Stats())
}

// ===== 货源排行 (对齐旧 ddtj.php) =====

func AdminSupplierRanking(c *gin.Context) {
	list, err := adminService.SupplierRanking()
	if err != nil {
		response.ServerError(c, "查询货源排行失败")
		return
	}
	response.Success(c, list)
}

// AdminRedockPending 将所有 dockstatus=0 的待对接订单重新推入队列
func AdminRedockPending(c *gin.Context) {
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "对接队列未初始化")
		return
	}
	rows, err := database.DB.Query("SELECT oid FROM qingka_wangke_order WHERE dockstatus = 0")
	if err != nil {
		response.ServerError(c, fmt.Sprintf("查询失败: %v", err))
		return
	}
	defer rows.Close()
	var oids []int64
	for rows.Next() {
		var oid int64
		rows.Scan(&oid)
		oids = append(oids, oid)
	}
	if len(oids) == 0 {
		response.Success(c, "无待对接订单")
		return
	}
	queue.GlobalDockQueue.PushBatch(oids)
	response.Success(c, fmt.Sprintf("已推入 %d 个订单", len(oids)))
}

// ===== 代理商品排行 (对齐旧 dl.php) =====

func AdminAgentProductRanking(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Query("uid"))
	timeType := c.DefaultQuery("time", "today")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := adminService.AgentProductRanking(uid, timeType, limit)
	if err != nil {
		response.ServerError(c, "查询代理商品排行失败")
		return
	}
	response.Success(c, list)
}
