package supplier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-api/internal/model"
)

func (s *Service) HasBatchProgressConfig(sup *model.SupplierFull) bool {
	if sup == nil {
		return false
	}
	return s.HasBatchProgressForPT(sup.PT)
}

func (s *Service) HasBatchProgressForPT(pt string) bool {
	cfg := GetPlatformConfig(pt)
	return hasExplicitActionConfig(cfg.BatchProgressPath, cfg.BatchProgressMethod, cfg.BatchProgressParamMap)
}

func (s *Service) QueryBatchOrderProgress(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error) {
	cfg := GetPlatformConfig(sup.PT)
	if err := requireExplicitActionConfig("批量进度接口", cfg.BatchProgressPath, cfg.BatchProgressMethod, cfg.BatchProgressParamMap); err != nil {
		return nil, err
	}

	apiURL := resolveConfiguredActionURL(sup.URL, cfg.BatchProgressPath)
	fallbackBodyType := "form"
	if strings.EqualFold(cfg.BatchProgressMethod, http.MethodGet) {
		fallbackBodyType = "query"
	}

	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.BatchProgressMethod,
		cfg.BatchProgressBodyType,
		cfg.BatchProgressParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultSupplierAuthParams(sup, cfg.AuthType),
		buildBatchProgressActionFields(refs),
	)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	return parseConfiguredProgressResponse(execResult.Body)
}

func buildBatchProgressActionFields(refs []model.SupplierBatchProgressRef) map[string]string {
	yids := make([]string, 0, len(refs))
	users := make([]string, 0, len(refs))
	kcnames := make([]string, 0, len(refs))
	kcids := make([]string, 0, len(refs))
	nouns := make([]string, 0, len(refs))
	seen := map[string]map[string]bool{
		"yid":    {},
		"user":   {},
		"kcname": {},
		"kcid":   {},
		"noun":   {},
	}

	appendUnique := func(bucket []string, kind string, value string) []string {
		value = strings.TrimSpace(value)
		if value == "" || seen[kind][value] {
			return bucket
		}
		seen[kind][value] = true
		return append(bucket, value)
	}

	for _, ref := range refs {
		yids = appendUnique(yids, "yid", ref.YID)
		users = appendUnique(users, "user", ref.User)
		kcnames = appendUnique(kcnames, "kcname", ref.KCName)
		kcids = appendUnique(kcids, "kcid", ref.KCID)
		nouns = appendUnique(nouns, "noun", ref.Noun)
	}

	return map[string]string{
		"batch.count":   strconv.Itoa(len(refs)),
		"batch.yids":    strings.Join(yids, ","),
		"batch.users":   strings.Join(users, ","),
		"batch.kcnames": strings.Join(kcnames, ","),
		"batch.kcids":   strings.Join(kcids, ","),
		"batch.nouns":   strings.Join(nouns, ","),
	}
}

func parseConfiguredProgressResponse(body []byte) ([]model.SupplierProgressItem, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	codeVal := ""
	if codeRaw, ok := raw["code"]; ok {
		switch v := codeRaw.(type) {
		case string:
			codeVal = v
		case float64:
			codeVal = fmt.Sprintf("%.0f", v)
		case int:
			codeVal = fmt.Sprintf("%d", v)
		default:
			codeVal = fmt.Sprintf("%v", v)
		}
	}
	if codeVal != "0" && codeVal != "1" {
		msg := toString(raw["msg"])
		if isNoBatchProgressUpdateMessage(msg) {
			return []model.SupplierProgressItem{}, nil
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	items := make([]model.SupplierProgressItem, 0)
	switch data := raw["data"].(type) {
	case []interface{}:
		for _, item := range data {
			row, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			items = append(items, supplierProgressItemFromMap(row))
		}
	case map[string]interface{}:
		items = append(items, supplierProgressItemFromMap(data))
	}

	return items, nil
}

func isNoBatchProgressUpdateMessage(msg string) bool {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return false
	}
	patterns := []string{
		"无可更新进度订单",
		"没有可更新进度订单",
		"暂无可更新进度订单",
		"无可更新订单",
	}
	for _, pattern := range patterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

func supplierProgressItemFromMap(data map[string]interface{}) model.SupplierProgressItem {
	firstValue := func(keys ...string) string {
		for _, key := range keys {
			if value := toString(data[key]); strings.TrimSpace(value) != "" {
				return value
			}
		}
		return ""
	}

	return model.SupplierProgressItem{
		YID:             firstValue("id", "yid", "uuid"),
		Noun:            firstValue("cid", "noun", "courseid", "course_id"),
		KCName:          firstValue("kcname", "course", "course_name", "courseName"),
		User:            firstValue("user", "username", "account"),
		Status:          firstValue("status", "state"),
		StatusText:      firstValue("status_text", "statusText", "status_msg"),
		Process:         firstValue("process", "progress"),
		Remarks:         firstValue("remarks", "message", "msg"),
		CourseStartTime: firstValue("courseStartTime", "kcks"),
		CourseEndTime:   firstValue("courseEndTime", "kcjs"),
		ExamStartTime:   firstValue("examStartTime", "ksks"),
		ExamEndTime:     firstValue("examEndTime", "ksjs"),
	}
}
