package sdxy

import (
	"fmt"

	"go-api/internal/database"
)

func (s *SDXYService) ProxyGetUserInfo(form map[string]interface{}) ([]byte, error) {
	params := map[string]string{}
	for k, v := range form {
		params["form["+k+"]"] = fmt.Sprintf("%v", v)
	}
	return s.upstreamRaw("get_user_info_by_password", params)
}

func (s *SDXYService) ProxySendCode(form map[string]interface{}) ([]byte, error) {
	params := map[string]string{}
	for k, v := range form {
		params["form["+k+"]"] = fmt.Sprintf("%v", v)
	}
	return s.upstreamRaw("send_code", params)
}

func (s *SDXYService) ProxyGetUserInfoByCode(form map[string]interface{}) ([]byte, error) {
	params := map[string]string{}
	for k, v := range form {
		params["form["+k+"]"] = fmt.Sprintf("%v", v)
	}
	return s.upstreamRaw("get_user_info_by_code", params)
}

func (s *SDXYService) ProxyUpdateRunRule(studentId string) ([]byte, error) {
	return s.upstreamRaw("update_run_rule", map[string]string{"student_id": studentId})
}

func (s *SDXYService) ProxyGetRunTask(uid int, sdxyOrderId string, pageNum, pageSize int, isAdmin bool) ([]byte, error) {
	var orderUID int
	err := database.DB.QueryRow(
		"SELECT uid FROM qingka_wangke_flash_sdxy WHERE sdxy_order_id = ? LIMIT 1", sdxyOrderId,
	).Scan(&orderUID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("您暂无权限")
	}
	return s.upstreamRaw("log", map[string]string{
		"sdxy_order_id": sdxyOrderId,
		"page_num":      fmt.Sprintf("%d", pageNum),
		"page_size":     fmt.Sprintf("%d", pageSize),
	})
}

func (s *SDXYService) ProxyChangeTaskTime(uid int, sdxyOrderId, runTaskId, startTime string, isAdmin bool) ([]byte, error) {
	var orderUID int
	err := database.DB.QueryRow(
		"SELECT uid FROM qingka_wangke_flash_sdxy WHERE sdxy_order_id = ? LIMIT 1", sdxyOrderId,
	).Scan(&orderUID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("您暂无权限")
	}
	return s.upstreamRaw("change_task_time", map[string]string{
		"sdxy_order_id": sdxyOrderId,
		"run_task_id":   runTaskId,
		"start_time":    startTime,
	})
}

func (s *SDXYService) ProxyDelayTask(uid int, aggOrderId, runTaskId string, isAdmin bool) ([]byte, error) {
	var orderUID int
	err := database.DB.QueryRow(
		"SELECT uid FROM qingka_wangke_flash_sdxy WHERE agg_order_id = ? LIMIT 1", aggOrderId,
	).Scan(&orderUID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("您暂无权限")
	}
	return s.upstreamRaw("delay_task", map[string]string{
		"agg_order_id": aggOrderId,
		"run_task_id":  runTaskId,
	})
}
