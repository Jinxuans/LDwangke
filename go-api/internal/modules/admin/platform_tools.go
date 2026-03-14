package admin

import (
	"go-api/internal/model"
	"go-api/internal/platformtools"
)

type DetectRequest = platformtools.DetectRequest
type ProbeDetail = platformtools.ProbeDetail
type DetectResult = platformtools.DetectResult

type AutoDetectRequest struct {
	DetectRequest
	PT   string `json:"pt" binding:"required"`
	Name string `json:"name"`
}

type ParsedPHPConfig = platformtools.ParsedPHPConfig

// 平台探测与 PHP 解析的真实 owner 在 platformtools，admin 只保留 HTTP 入参包装。
func adminDetectPlatform(req DetectRequest) *DetectResult {
	return platformtools.DetectPlatform(req)
}

func buildAdminConfigFromDetection(result *DetectResult, pt, name string) *model.PlatformConfigSaveRequest {
	if result == nil {
		return nil
	}
	return platformtools.BuildConfigFromDetection(result, pt, name)
}

func parseAdminPHPCode(code string) *ParsedPHPConfig {
	return platformtools.ParsePHPCode(code)
}
