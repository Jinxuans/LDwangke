package supplier

import "testing"

func TestParseActionJSONResponseAndHelpers(t *testing.T) {
	resp, err := parseActionJSONResponse([]byte(`{"code":1,"msg":"ok","data":[{"id":"7","name":"math"}],"userName":"alice"}`))
	if err != nil {
		t.Fatalf("parseActionJSONResponse returned error: %v", err)
	}
	if got := resp.code(); got != "1" {
		t.Fatalf("code = %q, want %q", got, "1")
	}
	if got := resp.msg(); got != "ok" {
		t.Fatalf("msg = %q, want %q", got, "ok")
	}
	rows := resp.dataRows()
	if len(rows) != 1 {
		t.Fatalf("dataRows len = %d, want 1", len(rows))
	}
	if got := firstActionValue(rows[0], "name", "course_name"); got != "math" {
		t.Fatalf("firstActionValue = %q, want %q", got, "math")
	}
	if got := resp.stringValue("userName"); got != "alice" {
		t.Fatalf("userName = %q, want %q", got, "alice")
	}
}
