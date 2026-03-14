package admin

import (
	"go-api/internal/autosync"
	"go-api/internal/dbtools"
	"go-api/internal/pluginruntime"
)

type SyncRequest = dbtools.SyncRequest
type SyncTableInfo = dbtools.SyncTableInfo
type SyncTestResult = dbtools.SyncTestResult
type SyncResult = dbtools.SyncResult
type MissingColumnInfo = dbtools.MissingColumnInfo
type CompatCheckResult = dbtools.CompatCheckResult
type CompatFixResult = dbtools.CompatFixResult

// 这些 helper 只服务于 admin 运维入口，把运行态工具调用集中在 admin 模块内部，
// 避免为了少量后台工具长期保留一个单独的顶级桥接目录。
func getAdminAutoSyncStatus() map[string]interface{} {
	return autosync.AutoSyncStatus()
}

func restartAdminHZWSocket() {
	pluginruntime.RestartHZWSocket()
}

func runAdminLonglongSyncOnce() (string, error) {
	return pluginruntime.LonglongSyncOnce()
}

func getAdminLonglongStatus() map[string]interface{} {
	return pluginruntime.LonglongStatus()
}

func getAdminLonglongCLIStatus() map[string]interface{} {
	return pluginruntime.LonglongCheckCLI()
}

func installAdminLonglongCLI() (string, error) {
	return pluginruntime.LonglongInstallCLI()
}

func checkAdminDBCompat() (*CompatCheckResult, error) {
	// db-compat 是后台运维工具的一部分，但真实 owner 在 dbtools；
	// admin 这里只保留一个模块内入口，避免 handler 直接散落到底层运行包。
	return dbtools.CheckDBCompat()
}

func fixAdminDBCompat() (*CompatFixResult, error) {
	// 修复动作直接交给 dbtools，避免继续引入额外桥接层。
	return dbtools.FixDBCompat()
}

func testAdminDBSyncConnection(req SyncRequest) (*SyncTestResult, error) {
	// 后台测试入口只做一次轻量探测，帮助管理员先确认外部库可连通、目标表存在。
	return dbtools.TestDBSyncConnection(req)
}

func executeAdminDBSync(req SyncRequest) (*SyncResult, error) {
	// 后台正式执行入口直接进入 dbtools，同步 owner 不再经过额外桥接层。
	return dbtools.ExecuteDBSync(req)
}
