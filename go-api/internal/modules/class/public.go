package class

import (
	"go-api/internal/database"
	"go-api/internal/model"
	suppliermodule "go-api/internal/modules/supplier"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var supplierService = suppliermodule.SharedService()

func getCategorySwitchesByCID(cid int) (log, ticket, changepass, allowpause, supplierReport, supplierReportHID int) {
	changepass = 1
	err := database.DB.QueryRow(
		`SELECT COALESCE(f.log,0), COALESCE(f.ticket,0), COALESCE(f.changepass,1), COALESCE(f.allowpause,0), COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		 FROM qingka_wangke_class c
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE c.cid = ?`, cid,
	).Scan(&log, &ticket, &changepass, &allowpause, &supplierReport, &supplierReportHID)
	if err != nil {
		return 0, 0, 1, 0, 0, 0
	}
	return
}

func listCategories() ([]model.Category, error) {
	rows, err := database.DB.Query(
		"SELECT id, name, sort, status, COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0) FROM qingka_wangke_fenlei WHERE status >= 1 ORDER BY recommend DESC, sort ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Sort, &c.Status, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []model.Category{}
	}
	return categories, nil
}

func List(c *gin.Context) {
	var req model.ClassListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = model.ClassListRequest{}
	}

	uid := c.GetInt("uid")
	list, err := classes.ListClasses(uid, req)
	if err != nil {
		response.ServerError(c, "查询课程失败")
		return
	}

	response.Success(c, list)
}

func ListPaged(c *gin.Context) {
	var req model.ClassListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = model.ClassListRequest{}
	}

	uid := c.GetInt("uid")
	list, total, page, limit, err := classes.ListClassesPaged(uid, req)
	if err != nil {
		response.ServerError(c, "查询课程失败")
		return
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"has_more": int64(page*limit) < total,
		},
	})
}

func Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "请输入搜索关键词")
		return
	}

	uid := c.GetInt("uid")
	req := model.ClassListRequest{Search: keyword}
	list, err := classes.ListClasses(uid, req)
	if err != nil {
		response.ServerError(c, "搜索课程失败")
		return
	}

	response.Success(c, list)
}

func QueryCourse(c *gin.Context) {
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

func CategorySwitches(c *gin.Context) {
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
	log, ticket, changepass, allowpause, supplierReport, supplierReportHID := getCategorySwitchesByCID(cidInt)
	response.Success(c, gin.H{
		"log":                 log,
		"ticket":              ticket,
		"changepass":          changepass,
		"allowpause":          allowpause,
		"supplier_report":     supplierReport,
		"supplier_report_hid": supplierReportHID,
	})
}

func Categories(c *gin.Context) {
	categories, err := listCategories()
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}

	response.Success(c, categories)
}
