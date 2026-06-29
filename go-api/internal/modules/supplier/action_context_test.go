package supplier

import (
	"testing"

	"go-api/internal/model"
)

func TestActionTemplateContextRendersOrderAndActionVariables(t *testing.T) {
	sup := &model.SupplierFull{User: "uid-1", Pass: "key-1", Token: "tok-1", Cookie: "cookie-1", URL: "https://example.test", PT: "demo"}
	ctx := newActionTemplateContext(sup, map[string]string{
		"order.user":      "alice",
		"order.pass":      "secret",
		"order.kcname":    "course-a",
		"action.school":   "school-a",
		"action.user":     "bob",
		"action.password": "pw-b",
		"action.platform": "plat-a",
	})

	rendered, err := ctx.render(`{"uid":"{{supplier.uid}}","token":"{{supplier.token}}","user":"{{order.user}}","pass":"{{order.pass}}","school":"{{action.school}}","platform":"{{action.platform}}"}`)
	if err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	expected := `{"uid":"uid-1","token":"key-1","user":"alice","pass":"secret","school":"school-a","platform":"plat-a"}`
	if rendered != expected {
		t.Fatalf("unexpected render\nexpected: %s\nactual:   %s", expected, rendered)
	}
}

func TestOrderProgressActionFieldsSkipsDebugKeys(t *testing.T) {
	fields := orderProgressActionFields("yid-1", "alice", map[string]string{
		"kcname":       "course-a",
		"pass":         "secret",
		"__debug_http": "1",
		"__debug_oid":  "123",
	})

	if got := fields["order.yid"]; got != "yid-1" {
		t.Fatalf("order.yid = %q, want %q", got, "yid-1")
	}
	if got := fields["order.user"]; got != "alice" {
		t.Fatalf("order.user = %q, want %q", got, "alice")
	}
	if got := fields["order.pass"]; got != "secret" {
		t.Fatalf("order.pass = %q, want %q", got, "secret")
	}
	if _, ok := fields["order.__debug_http"]; ok {
		t.Fatalf("debug key should not be copied into order.*")
	}
}
