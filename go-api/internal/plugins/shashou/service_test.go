package shashou

import "testing"

func TestExtractUpstreamOrderID(t *testing.T) {
	payload := map[string]any{
		"code": 200,
		"data": map[string]any{
			"order_id": "12345",
		},
	}
	if got := extractUpstreamOrderID(payload); got != "12345" {
		t.Fatalf("extractUpstreamOrderID() = %q, want 12345", got)
	}
	payload["upstream_order_id"] = "67890"
	if got := extractUpstreamOrderID(payload); got != "67890" {
		t.Fatalf("extractUpstreamOrderID() = %q, want stored upstream id", got)
	}
}

func TestExtractUpstreamOrderNoPrefersNestedOrderNo(t *testing.T) {
	payload := map[string]any{
		"id": "99",
		"data": map[string]any{
			"order_no": "SS202605170001",
			"id":       "100",
		},
	}
	if got := extractUpstreamOrderNo(payload); got != "SS202605170001" {
		t.Fatalf("extractUpstreamOrderNo() = %q, want nested order_no", got)
	}
}

func TestRefundResultParsing(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"refund_result": map[string]any{
				"status":        "success",
				"refund_amount": "4.567",
			},
		},
	}
	if !refundSucceeded(payload) {
		t.Fatal("refundSucceeded() = false, want true")
	}
	if got := refundAmount(payload); got != 4.57 {
		t.Fatalf("refundAmount() = %.2f, want 4.57", got)
	}
}

func TestAggregateDetailStatusFailed(t *testing.T) {
	payload := map[string]any{
		"msg": "同步成功",
		"data": map[string]any{
			"detail": map[string]any{
				"accounts": []any{
					map[string]any{
						"account":       "taixuex.cc",
						"status":        "failed",
						"error_message": "用户名或密码错误",
					},
				},
			},
			"status":      "completed",
			"actual_cost": 0,
		},
	}
	if got := aggregateDetailStatus(payload); got != "failed" {
		t.Fatalf("aggregateDetailStatus() = %q, want failed", got)
	}
	if got := orderSyncError(payload, "failed"); got != "用户名或密码错误" {
		t.Fatalf("orderSyncError() = %q, want account error", got)
	}
	order := Order{Status: "pending"}
	if got := localSyncStatus(order, payload); got != "failed" {
		t.Fatalf("localSyncStatus() = %q, want failed", got)
	}
}

func TestCanSettlePayment(t *testing.T) {
	for _, status := range []string{"", "pre_deducted", "partial_refund", "insufficient"} {
		if !canSettlePayment(status) {
			t.Fatalf("canSettlePayment(%q) = false, want true", status)
		}
	}
	for _, status := range []string{"settled", "refunded", "no_refund"} {
		if canSettlePayment(status) {
			t.Fatalf("canSettlePayment(%q) = true, want false", status)
		}
	}
}

func TestMapStatusPendingAliases(t *testing.T) {
	for _, raw := range []string{"pending", "待处理", "1"} {
		if got := mapStatus(raw); got != "pending" {
			t.Fatalf("mapStatus(%q) = %q, want pending", raw, got)
		}
	}
	if got := mapStatus("2"); got != "processing" {
		t.Fatalf("mapStatus(%q) = %q, want processing", "2", got)
	}
	if got := mapStatus("3"); got != "completed" {
		t.Fatalf("mapStatus(%q) = %q, want completed", "3", got)
	}
}

func TestActualSettlementOnlyOnTerminalStatus(t *testing.T) {
	for _, status := range []string{"pending", "processing"} {
		if canApplyActualSettlement(status) {
			t.Fatalf("canApplyActualSettlement(%q) = true, want false", status)
		}
	}
	for _, status := range []string{"completed", "failed"} {
		if !canApplyActualSettlement(status) {
			t.Fatalf("canApplyActualSettlement(%q) = false, want true", status)
		}
	}
}

func TestPendingPayloadKeepsLocalPending(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"status":      "processing",
			"actual_cost": 0,
			"detail": map[string]any{
				"accounts": []any{
					map[string]any{
						"account": "taixuex.cc",
						"status":  "1",
					},
				},
			},
		},
	}
	order := Order{Status: "pending"}
	if got := localSyncStatus(order, payload); got != "pending" {
		t.Fatalf("localSyncStatus() = %q, want pending", got)
	}
}
