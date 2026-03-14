package service

import "go-api/internal/model"

type AdminService struct{}

var adminService = &AdminService{}

func GetAdminConfigMap() (map[string]string, error) {
	return adminService.GetConfig()
}

func DashboardStatsData() (map[string]interface{}, error) {
	return adminService.DashboardStats()
}

func StatsReportData(days int) (map[string]interface{}, error) {
	return adminService.StatsReport(days)
}

func AdminMoneyLogListData(page, limit int, uid string, logType string) ([]map[string]interface{}, int64, error) {
	return adminService.AdminMoneyLogList(page, limit, uid, logType)
}

func SupplierRankingData() ([]SupplierRankItem, error) {
	return adminService.SupplierRanking()
}

func AgentProductRankingData(uid int, timeType string, limit int) ([]AgentProductRankItem, error) {
	return adminService.AgentProductRanking(uid, timeType, limit)
}

func UserListData(req model.UserListRequest) ([]model.UserManage, int64, error) {
	return adminService.UserList(req)
}

func UserResetPasswordData(uid int, newPass string) error {
	return adminService.UserResetPassword(uid, newPass)
}

func UserSetBalanceData(uid int, balance float64) error {
	return adminService.UserSetBalance(uid, balance)
}

func UserSetGradeData(uid int, addprice float64) error {
	return adminService.UserSetGrade(uid, addprice)
}

func GradeListData() ([]model.Grade, error) {
	return adminService.GradeList()
}

func GradeSaveData(req model.GradeSaveRequest) error {
	return adminService.GradeSave(req)
}

func GradeDeleteData(id int) error {
	return adminService.GradeDelete(id)
}

func SupplierListData() ([]model.Supplier, error) {
	return adminService.SupplierList()
}

func SupplierSaveData(sup model.Supplier) error {
	return adminService.SupplierSave(sup)
}

func SupplierDeleteData(hid int) error {
	return adminService.SupplierDelete(hid)
}

func CategoryListAllData() ([]model.Category, error) {
	return adminService.CategoryListAll()
}

func CategoryListPagedData(req model.CategoryListRequest) ([]model.Category, int64, error) {
	return adminService.CategoryListPaged(req)
}

func CategorySaveData(cat model.Category) error {
	return adminService.CategorySave(cat)
}

func CategoryDeleteData(id int) error {
	return adminService.CategoryDelete(id)
}

func CategoryQuickModifyData(keyword string, categoryID int) (int64, error) {
	return adminService.CategoryQuickModify(keyword, categoryID)
}

func CategoryUpdateSortData(items []struct{ ID, Sort int }) error {
	return adminService.CategoryUpdateSort(items)
}

func ClassListData(cateID int, keywords string, page, limit int) ([]model.ClassManage, int64, error) {
	return adminService.ClassList(cateID, keywords, page, limit)
}

func ClassSaveData(req model.ClassEditRequest) error {
	return adminService.ClassSave(req)
}

func ClassToggleStatusData(cid, status int) error {
	return adminService.ClassToggleStatus(cid, status)
}

func ClassBatchDeleteData(cids []int) (int64, error) {
	return adminService.ClassBatchDelete(cids)
}

func ClassBatchCategoryData(cids []int, cateID string) (int64, error) {
	return adminService.ClassBatchCategory(cids, cateID)
}

func ClassBatchPriceData(cids []int, rate float64, yunsuan string) (int64, error) {
	return adminService.ClassBatchPrice(cids, rate, yunsuan)
}

func ClassBatchReplaceKeywordData(search, replace, scope, scopeID string) (int64, error) {
	return adminService.ClassBatchReplaceKeyword(search, replace, scope, scopeID)
}

func ClassBatchAddPrefixData(prefix, scope, scopeID string) (int64, error) {
	return adminService.ClassBatchAddPrefix(prefix, scope, scopeID)
}

func MiJiaListData(req model.MiJiaListRequest) ([]model.MiJia, int64, []int, error) {
	return adminService.MiJiaList(req)
}

func MiJiaSaveData(req model.MiJiaSaveRequest) error {
	return adminService.MiJiaSave(req)
}

func MiJiaDeleteData(mids []int) error {
	return adminService.MiJiaDelete(mids)
}

func MiJiaBatchData(req model.MiJiaBatchRequest) (int, error) {
	return adminService.MiJiaBatch(req)
}
