package platformtools

import (
	"strings"
	"testing"
)

func TestBuildConfigFromDetection(t *testing.T) {
	result := &DetectResult{
		Success:       true,
		SuggestedName: "demo-platform",
		ReturnsYID:    true,
		Config: map[string]string{
			"auth_type":           "token_only",
			"success_codes":       "0,1",
			"use_json":            "true",
			"query_path":          "/api/get",
			"order_path":          "/api/add",
			"progress_path":       "/api/query",
			"progress_method":     "POST",
			"pause_path":          "/api/stop",
			"change_pass_path":    "/api/update",
			"report_path":         "/api/submitWork",
			"get_report_path":     "/api/queryWork",
			"report_param_style":  "token",
			"balance_money_field": "data.balance",
		},
	}

	cfg := BuildConfigFromDetection(result, "demo", "")
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Name != "demo-platform" {
		t.Fatalf("expected suggested name, got %q", cfg.Name)
	}
	if cfg.AuthType != "token_only" {
		t.Fatalf("unexpected auth type: %+v", cfg)
	}
	if cfg.QueryPath != "/api/get" || cfg.OrderPath != "/api/add" {
		t.Fatalf("unexpected query/order path: %+v", cfg)
	}
	if !cfg.UseJSON || !cfg.ReturnsYID {
		t.Fatalf("expected flags to be enabled: %+v", cfg)
	}
	if cfg.BalanceMoneyField != "data.balance" {
		t.Fatalf("unexpected balance field: %q", cfg.BalanceMoneyField)
	}
	if cfg.ProgressParamMap != "" {
		t.Fatalf("expected empty progress param map, got %q", cfg.ProgressParamMap)
	}
}

func TestParsePHPCodeExtractsSignals(t *testing.T) {
	code := `
	if ($type == "xm") {
		$url = "/api/query";
		$data = ["token" => $a["pass"], "id" => $d["yid"]];
		if ($result['code'] == 0 || $result['code'] == 1 || $result['code'] == 200) {
			$ok = true;
		}
		$payload = ["newPwd" => "123456", "id" => $d["yid"]];
	}
	`

	cfg := ParsePHPCode(code)
	if cfg == nil {
		t.Fatal("expected parsed config")
	}
	if cfg.PT != "xm" {
		t.Fatalf("expected pt xm, got %q", cfg.PT)
	}
	if cfg.AuthType != "token_only" {
		t.Fatalf("expected token_only auth, got %q", cfg.AuthType)
	}
	if cfg.QueryPath != "/api/query" {
		t.Fatalf("unexpected parsing result: %+v", cfg)
	}
	if !strings.Contains(cfg.SuccessCodes, "0") || !strings.Contains(cfg.SuccessCodes, "200") {
		t.Fatalf("unexpected success codes: %q", cfg.SuccessCodes)
	}
	if cfg.ChangePassParam != "newPwd" || cfg.ChangePassID != "id" {
		t.Fatalf("unexpected change pass params: %+v", cfg)
	}
	if cfg.Confidence <= 0 {
		t.Fatalf("expected positive confidence, got %d", cfg.Confidence)
	}
}

func TestParseProbeDetailAndExtractMoney(t *testing.T) {
	detail := parseProbeDetail(ProbeDetail{Endpoint: "balance"}, []byte(`{"code":0,"data":{"balance":15.5}}`), false, 200)
	if detail.Status != "ok" || detail.Code != "0" {
		t.Fatalf("unexpected probe detail: %+v", detail)
	}

	money := tryExtractMoney(map[string]interface{}{
		"data": map[string]interface{}{"balance": 15.5},
	})
	if money.field != "data.balance" {
		t.Fatalf("unexpected money field: %+v", money)
	}
	if money.value == "" {
		t.Fatal("expected money value")
	}
}
