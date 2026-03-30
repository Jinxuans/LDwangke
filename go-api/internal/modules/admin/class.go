package admin

import (
	"fmt"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerClassRoutes(admin *gin.RouterGroup) {
	admin.GET("/categories", AdminCategoryList)
	admin.GET("/categories/paged", AdminCategoryListPaged)
	admin.POST("/category/save", AdminCategorySave)
	admin.DELETE("/category/:id", AdminCategoryDelete)
	admin.POST("/category/quick-modify", AdminCategoryQuickModify)
	admin.POST("/category/update-sort", AdminCategoryUpdateSort)
	admin.POST("/category/batch-toggle", AdminCategoryBatchToggle)

	admin.GET("/classes", AdminClassList)
	admin.POST("/class/save", AdminClassSave)
	admin.POST("/class/toggle", AdminClassToggle)
	admin.POST("/class/batch-delete", AdminClassBatchDelete)
	admin.POST("/class/batch-category", AdminClassBatchCategory)
	admin.POST("/class/batch-price", AdminClassBatchPrice)
	admin.POST("/class/batch-replace-keyword", AdminClassBatchReplaceKeyword)
	admin.POST("/class/batch-add-prefix", AdminClassBatchAddPrefix)
	admin.POST("/class/add", AdminAddClass)
	admin.GET("/class/dropdown", AdminClassDropdown)

	admin.GET("/mijia", AdminMiJiaList)
	admin.POST("/mijia/save", AdminMiJiaSave)
	admin.POST("/mijia/delete", AdminMiJiaDelete)
	admin.POST("/mijia/batch", AdminMiJiaBatch)
}

func AdminCategoryList(c *gin.Context) {
	list, err := classmodule.Classes().CategoryListAll()
	if err != nil {
		response.ServerErrorf(c, err, "查询分类失败")
		return
	}
	response.Success(c, list)
}

func AdminCategoryListPaged(c *gin.Context) {
	var req model.CategoryListRequest
	_ = c.ShouldBindQuery(&req)

	list, total, err := classmodule.Classes().CategoryListPaged(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询分类失败")
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
	if err := classmodule.Classes().CategorySave(cat); err != nil {
		response.ServerErrorf(c, err, "保存分类失败")
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
	if err := classmodule.Classes().CategoryDelete(id); err != nil {
		response.ServerErrorf(c, err, "删除分类失败")
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
	affected, err := classmodule.Classes().CategoryQuickModify(body.Keyword, body.CategoryID)
	if err != nil {
		response.ServerErrorf(c, err, "修改失败")
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

	var items []struct{ ID, Sort int }
	for _, item := range body.Items {
		items = append(items, struct{ ID, Sort int }{ID: item.ID, Sort: item.Sort})
	}
	if err := classmodule.Classes().CategoryUpdateSort(items); err != nil {
		response.ServerErrorf(c, err, "排序更新失败")
		return
	}
	response.Success(c, gin.H{"msg": "排序更新成功"})
}

func AdminClassList(c *gin.Context) {
	cateID, _ := strconv.Atoi(c.Query("cateId"))
	keywords := c.Query("keywords")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	list, total, err := classmodule.Classes().ClassList(cateID, keywords, page, limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询课程失败")
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
	if err := classmodule.Classes().ClassSave(req); err != nil {
		response.ServerErrorf(c, err, "保存课程失败")
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
	if err := classmodule.Classes().ClassToggleStatus(body.CID, body.Status); err != nil {
		response.ServerErrorf(c, err, "更新状态失败")
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
	deleted, err := classmodule.Classes().ClassBatchDelete(body.CIDs)
	if err != nil {
		response.ServerErrorf(c, err, "删除失败: "+err.Error())
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
	updated, err := classmodule.Classes().ClassBatchCategory(body.CIDs, body.CateID)
	if err != nil {
		response.ServerErrorf(c, err, "批量修改分类失败")
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
	updated, err := classmodule.Classes().ClassBatchPrice(body.CIDs, body.Rate, body.Yunsuan)
	if err != nil {
		response.ServerErrorf(c, err, "批量修改价格失败")
		return
	}
	response.Success(c, gin.H{"updated": updated, "msg": fmt.Sprintf("成功修改 %d 个课程的价格", updated)})
}

func AdminClassBatchReplaceKeyword(c *gin.Context) {
	var body struct {
		Search  string `json:"search"`
		Replace string `json:"replace"`
		Scope   string `json:"scope"`
		ScopeID string `json:"scope_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.Search == "" {
		response.BadRequest(c, "请输入要替换的关键词")
		return
	}
	if body.Scope == "" {
		body.Scope = "all"
	}
	updated, err := classmodule.Classes().ClassBatchReplaceKeyword(body.Search, body.Replace, body.Scope, body.ScopeID)
	if err != nil {
		response.ServerErrorf(c, err, "替换失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated, "msg": fmt.Sprintf("成功替换 %d 个课程的关键词", updated)})
}

func AdminClassBatchAddPrefix(c *gin.Context) {
	var body struct {
		Prefix  string `json:"prefix"`
		Scope   string `json:"scope"`
		ScopeID string `json:"scope_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.Prefix == "" {
		response.BadRequest(c, "请输入要添加的前缀")
		return
	}
	if body.Scope == "" {
		body.Scope = "all"
	}
	updated, err := classmodule.Classes().ClassBatchAddPrefix(body.Prefix, body.Scope, body.ScopeID)
	if err != nil {
		response.ServerErrorf(c, err, "添加前缀失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated, "msg": fmt.Sprintf("成功为 %d 个课程添加前缀", updated)})
}

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

	if req.Docking != "" && req.Docking != "0" && req.Noun != "" {
		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE docking=? AND noun=?", req.Docking, req.Noun).Scan(&cnt)
		if cnt > 0 {
			_, err := database.DB.Exec(
				"UPDATE qingka_wangke_class SET name=?, price=?, getnoun=?, content=?, queryplat=?, yunsuan=?, status=?, sort=?, fenlei=? WHERE docking=? AND noun=?",
				req.Name, req.Price, req.GetNoun, req.Content, req.QueryPlat, req.Yunsuan, statusVal, sortVal, req.Fenlei, req.Docking, req.Noun,
			)
			if err != nil {
				response.ServerErrorf(c, err, "更新失败: "+err.Error())
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
		response.ServerErrorf(c, err, "添加失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "添加成功")
}

func AdminCategoryBatchToggle(c *gin.Context) {
	var body struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	sourceStatus, targetStatus := 1, 1
	if body.Action == "down" {
		sourceStatus = 0
		targetStatus = 0
	}

	rows, err := database.DB.Query("SELECT id FROM qingka_wangke_fenlei WHERE status = ?", sourceStatus)
	if err != nil {
		response.ServerErrorf(c, err, "查询分类失败")
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

func AdminClassDropdown(c *gin.Context) {
	rows, err := database.DB.Query("SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(fenlei,'') FROM qingka_wangke_class WHERE status >= 0 ORDER BY sort ASC, cid ASC")
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	defer rows.Close()

	type classOption struct {
		CID    int    `json:"cid"`
		Name   string `json:"name"`
		Price  string `json:"price"`
		Fenlei string `json:"fenlei"`
	}

	var list []classOption
	for rows.Next() {
		var opt classOption
		rows.Scan(&opt.CID, &opt.Name, &opt.Price, &opt.Fenlei)
		list = append(list, opt)
	}
	if list == nil {
		list = []classOption{}
	}
	response.Success(c, list)
}

func AdminMiJiaList(c *gin.Context) {
	var req model.MiJiaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, uids, err := classmodule.Classes().MiJiaList(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询密价列表失败")
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
	if err := classmodule.Classes().MiJiaSave(req); err != nil {
		response.ServerErrorf(c, err, "保存失败")
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
	if err := classmodule.Classes().MiJiaDelete(req.Mids); err != nil {
		response.ServerErrorf(c, err, "删除失败")
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
	count, err := classmodule.Classes().MiJiaBatch(req)
	if err != nil {
		response.ServerErrorf(c, err, "批量设置失败")
		return
	}
	response.Success(c, gin.H{"count": count, "msg": fmt.Sprintf("成功为 %d 个商品设置密价", count)})
}
