package supplier

import (
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
	resp, err := parseActionJSONResponse(body)
	if err != nil {
		return nil, err
	}

	codeVal := resp.code()
	if codeVal != "0" && codeVal != "1" {
		msg := resp.msg()
		if isNoBatchProgressUpdateMessage(msg) {
			return []model.SupplierProgressItem{}, nil
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	items := make([]model.SupplierProgressItem, 0)
	for _, row := range resp.dataRows() {
		items = append(items, supplierProgressItemFromMap(row))
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
	return model.SupplierProgressItem{
		YID:             firstActionValue(data, "id", "yid", "uuid"),
		Noun:            firstActionValue(data, "cid", "noun", "courseid", "course_id"),
		KCName:          firstActionValue(data, "kcname", "course", "course_name", "courseName"),
		User:            firstActionValue(data, "user", "username", "account"),
		Status:          firstActionValue(data, "status", "state"),
		StatusText:      firstActionValue(data, "status_text", "statusText", "status_msg"),
		Process:         firstActionValue(data, "process", "progress"),
		Remarks:         firstActionValue(data, "remarks", "message", "msg"),
		CourseStartTime: firstActionValue(data, "courseStartTime", "kcks"),
		CourseEndTime:   firstActionValue(data, "courseEndTime", "kcjs"),
		ExamStartTime:   firstActionValue(data, "examStartTime", "ksks"),
		ExamEndTime:     firstActionValue(data, "examEndTime", "ksjs"),
	}
}
