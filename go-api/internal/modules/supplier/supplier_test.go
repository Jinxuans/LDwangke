package supplier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-api/internal/model"
)

func withRegistryPlatformConfigs(t *testing.T) {
	t.Helper()

	dbConfigMu.Lock()
	oldLoaded := dbConfigLoaded
	oldConfigCache := dbConfigCache
	oldNameCache := dbNameCache
	dbConfigLoaded = true
	dbConfigCache = map[string]PlatformConfig{
		"liunian": {
			AuthType:          "uid_key",
			SuccessCode:       "0",
			ReportPath:        "/api.php?act=submitWorkOrder",
			GetReportPath:     "/api.php?act=queryWorkOrder",
			ReportMethod:      "POST",
			GetReportMethod:   "POST",
			ReportParamStyle:  "standard",
			ReportParamMap:    `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","question":"{{action.content}}"}`,
			GetReportParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","reportId":"{{action.report_id}}"}`,
		},
		"2xx": {
			AuthType:          "uid_key",
			SuccessCode:       "1",
			ReportPath:        "/api/submitWork",
			GetReportPath:     "/api/queryWork",
			ReportMethod:      "POST",
			GetReportMethod:   "POST",
			ReportParamStyle:  "token",
			ReportAuthType:    "token_only",
			UseJSON:           true,
			ReportParamMap:    `{"token":"{{supplier.token}}","type":"{{action.ticket_type}}","id":"{{order.yid}}","content":"{{action.content}}"}`,
			GetReportParamMap: `{"token":"{{supplier.token}}","workId":"{{action.report_id}}"}`,
		},
	}
	dbNameCache = map[string]string{
		"liunian": "流年",
		"2xx":     "爱学习",
	}
	dbConfigMu.Unlock()

	t.Cleanup(func() {
		dbConfigMu.Lock()
		dbConfigLoaded = oldLoaded
		dbConfigCache = oldConfigCache
		dbNameCache = oldNameCache
		dbConfigMu.Unlock()
	})
}

func TestFillDefaults(t *testing.T) {
	cfg := fillDefaults(PlatformConfig{})

	if cfg.AuthType != "uid_key" {
		t.Fatalf("expected default AuthType=uid_key, got %q", cfg.AuthType)
	}
	if cfg.QueryPath != "" {
		t.Fatalf("expected QueryPath to remain explicit-only, got %q", cfg.QueryPath)
	}
	if cfg.QueryMethod != "" {
		t.Fatalf("expected QueryMethod to remain explicit-only, got %q", cfg.QueryMethod)
	}
	if cfg.OrderPath != "" {
		t.Fatalf("expected OrderPath to remain explicit-only, got %q", cfg.OrderPath)
	}
	if cfg.ProgressPath != "" {
		t.Fatalf("expected ProgressPath to remain explicit-only, got %q", cfg.ProgressPath)
	}
	if cfg.BalancePath != "" {
		t.Fatalf("expected BalancePath to remain explicit-only, got %q", cfg.BalancePath)
	}
	if cfg.ReportSuccessCode != "1" {
		t.Fatalf("expected default ReportSuccessCode=1, got %q", cfg.ReportSuccessCode)
	}
	if cfg.CategoryPath != "" || cfg.ClassListPath != "" {
		t.Fatalf("expected catalog paths to remain explicit-only: %+v", cfg)
	}
}

func TestParseCategoryResponse(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"data": []map[string]interface{}{
			{"fid": "10", "fname": "分类A"},
			{"category_id": 20, "category_name": "分类B"},
		},
	})

	got := parseCategoryResponse(body)
	if got["10"] != "分类A" {
		t.Fatalf("expected fid 10 => 分类A, got %q", got["10"])
	}
	if got["20"] != "分类B" {
		t.Fatalf("expected fid 20 => 分类B, got %q", got["20"])
	}
}

func TestExtractMoneyField(t *testing.T) {
	raw := map[string]interface{}{
		"money": "12.34",
		"data": map[string]interface{}{
			"money":       "56.78",
			"remainscore": 90,
		},
	}

	if got := extractMoneyField(raw, "money"); got != "12.34" {
		t.Fatalf("expected root money, got %q", got)
	}
	if got := extractMoneyField(raw, "data.money"); got != "56.78" {
		t.Fatalf("expected nested money, got %q", got)
	}
	if got := extractMoneyField(raw, "data.remainscore"); got != "90" {
		t.Fatalf("expected nested remainscore, got %q", got)
	}
}

func TestBuildSupplierURL(t *testing.T) {
	got := buildSupplierURL("example.com/", "get")
	want := "http://example.com/api.php?act=get"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveConfiguredActionURLPrefersPath(t *testing.T) {
	got := resolveConfiguredActionURL("example.com/", "/api/search")
	if got != "http://example.com/api/search" {
		t.Fatalf("unexpected resolved path url: %q", got)
	}

	got = resolveConfiguredActionURL("example.com/", "api.php?act=chadan")
	if got != "http://example.com/api.php?act=chadan" {
		t.Fatalf("unexpected resolved configured url: %q", got)
	}
}

func TestGetSupplierToken(t *testing.T) {
	sup := &model.SupplierFull{Pass: "pass-token", Token: "fallback-token"}
	if got := getSupplierToken(sup); got != "pass-token" {
		t.Fatalf("expected pass-token, got %q", got)
	}

	sup = &model.SupplierFull{Token: "fallback-token"}
	if got := getSupplierToken(sup); got != "fallback-token" {
		t.Fatalf("expected fallback-token, got %q", got)
	}
}

func TestUnsupportedPlatformOperations(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name string
		run  func() (string, error)
		want string
	}{
		{
			name: "pause yyy",
			run: func() (string, error) {
				_, msg, err := svc.PauseOrder(&model.SupplierFull{PT: "yyy"}, "1")
				return msg, err
			},
			want: "当前平台暂不支持暂停操作",
		},
		{
			name: "change password tuboshu",
			run: func() (string, error) {
				_, msg, err := svc.ChangePassword(&model.SupplierFull{PT: "tuboshu"}, "1", "new")
				return msg, err
			},
			want: "当前平台暂不支持改密操作",
		},
		{
			name: "logs yyy",
			run: func() (string, error) {
				_, err := svc.QueryOrderLogs(&model.SupplierFull{PT: "yyy"}, "1")
				if err == nil {
					return "", nil
				}
				return err.Error(), err
			},
			want: "当前平台暂不支持查看日志",
		},
		{
			name: "resubmit tuboshu",
			run: func() (string, error) {
				_, msg, err := svc.ResubmitOrder(&model.SupplierFull{PT: "tuboshu"}, "1")
				return msg, err
			},
			want: "当前平台暂不支持补单操作",
		},
		{
			name: "resubmit nx",
			run: func() (string, error) {
				_, msg, err := svc.ResubmitOrder(&model.SupplierFull{PT: "nx"}, "1")
				return msg, err
			},
			want: "当前平台暂不支持补单操作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			if err == nil && got == "" {
				t.Fatalf("expected unsupported error/message")
			}
			if got != tt.want {
				t.Fatalf("unexpected result: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTuboshuBalanceHelpers(t *testing.T) {
	if got := tuboshuBalanceAPIURL("demo.example.com"); got != "https://demo.example.com/api/userInfo" {
		t.Fatalf("unexpected tuboshu api url: %q", got)
	}
	if got := tuboshuBalanceAPIURL("https://demo.example.com/api"); got != "https://demo.example.com/api/userInfo" {
		t.Fatalf("unexpected tuboshu api url with api suffix: %q", got)
	}

	raw := map[string]interface{}{
		"data": map[string]interface{}{
			"point": 88.5,
		},
	}
	if got := extractTuboshuPoint(raw); got != "88.5" {
		t.Fatalf("unexpected tuboshu point: %q", got)
	}
	if got := extractTuboshuPoint(map[string]interface{}{}); got != "0" {
		t.Fatalf("expected default tuboshu point 0, got %q", got)
	}
}

func TestYYYHelpers(t *testing.T) {
	if got := yyyBaseURL("demo.example.com/"); got != "http://demo.example.com" {
		t.Fatalf("unexpected yyy base url: %q", got)
	}
	if got := yyyBaseURL("https://demo.example.com"); got != "https://demo.example.com" {
		t.Fatalf("unexpected yyy base url with scheme: %q", got)
	}

	raw := map[string]interface{}{
		"data": map[string]interface{}{
			"money": "123.45",
		},
	}
	if got := extractYYYMoney(raw); got != "123.45" {
		t.Fatalf("unexpected yyy money: %q", got)
	}
	if got := extractYYYMoney(map[string]interface{}{"money": 88}); got != "88" {
		t.Fatalf("unexpected fallback yyy money: %q", got)
	}
}

func TestSimpleHelpers(t *testing.T) {
	sup := &model.SupplierFull{URL: "demo.example.com/", Pass: "pass-token", Token: "token-value"}
	if got := simpleBuildBaseURL(sup); got != "http://demo.example.com" {
		t.Fatalf("unexpected simple base url: %q", got)
	}
	if got := simpleGetToken(sup); got != "token-value" {
		t.Fatalf("unexpected simple token: %q", got)
	}

	sup.Token = ""
	if got := simpleGetToken(sup); got != "pass-token" {
		t.Fatalf("unexpected simple fallback token: %q", got)
	}
}

func TestKunHelpers(t *testing.T) {
	sup := &model.SupplierFull{URL: "demo.example.com/", Pass: "pass-token", Token: "token-value"}
	if got := kunBuildBaseURL(sup); got != "http://demo.example.com" {
		t.Fatalf("unexpected kun base url: %q", got)
	}
	if got := kunGetDToken(sup); got != "token-value" {
		t.Fatalf("unexpected kun token: %q", got)
	}

	sup.Token = ""
	if got := kunGetDToken(sup); got != "pass-token" {
		t.Fatalf("unexpected kun fallback token: %q", got)
	}
}

func TestSimpleCallQueryParsesNestedChildren(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Api/Get" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("token") != "token-value" {
			t.Fatalf("unexpected token: %q", r.Form.Get("token"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    1,
			"student": "张三",
			"children": []map[string]interface{}{
				{
					"id":          "p1",
					"real_course": "父课程",
					"children": []map[string]interface{}{
						{"id": "c1", "real_course": "课程1"},
						{"id": "c2", "real_course": "课程2"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	result, err := simpleCallQuery(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "noun", "school", "user", "pass")
	if err != nil {
		t.Fatalf("simpleCallQuery returned error: %v", err)
	}
	if result.UserName != "张三" {
		t.Fatalf("unexpected username: %q", result.UserName)
	}
	if len(result.Data) != 2 || result.Data[0].ID != "c1" || result.Data[1].ID != "c2" {
		t.Fatalf("unexpected query data: %+v", result.Data)
	}
}

func TestSimpleResubmitUsesRepeatEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Api/Repeat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("token") != "token-value" || r.Form.Get("id") != "OID-1" {
			t.Fatalf("unexpected form: %#v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "补刷成功",
		})
	}))
	defer srv.Close()

	code, msg, err := simpleResubmit(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "OID-1")
	if err != nil {
		t.Fatalf("simpleResubmit returned error: %v", err)
	}
	if code != 1 || msg != "补刷成功" {
		t.Fatalf("unexpected resubmit result: code=%d msg=%q", code, msg)
	}
}

func TestKunCallOrderExtractsYID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/getorder/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("dtoken") != "token-value" || q.Get("platform") != "noun" || q.Get("course") != "课程A" {
			t.Fatalf("unexpected query: %#v", q)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  0,
			"msg":   "ok",
			"token": "YID-123",
		})
	}))
	defer srv.Close()

	result, err := kunCallOrder(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "noun", "school", "user", "pass", "课程A", "KC-1")
	if err != nil {
		t.Fatalf("kunCallOrder returned error: %v", err)
	}
	if result.Code != 1 || result.YID != "YID-123" {
		t.Fatalf("unexpected kun order result: %+v", result)
	}
}

func TestKunPauseAndChangePassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/uporder/":
			q := r.URL.Query()
			if q.Get("token") != "OID-1" || q.Get("state") != "暂停" || q.Get("dtoken") != "token-value" {
				t.Fatalf("unexpected pause query: %#v", q)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 1, "msg": "暂停成功"})
		case "/upPwd/":
			q := r.URL.Query()
			if q.Get("token") != "OID-1" || q.Get("pwd") != "new-pass" || q.Get("dtoken") != "token-value" {
				t.Fatalf("unexpected change password query: %#v", q)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 1, "msg": "改密成功"})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	code, msg, err := kunPauseOrder(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "OID-1")
	if err != nil || code != 1 || msg != "暂停成功" {
		t.Fatalf("unexpected kun pause result: code=%d msg=%q err=%v", code, msg, err)
	}

	code, msg, err = kunChangePassword(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "OID-1", "new-pass")
	if err != nil || code != 1 || msg != "改密成功" {
		t.Fatalf("unexpected kun change password result: code=%d msg=%q err=%v", code, msg, err)
	}
}

func TestKunCallQueryParsesCourses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/query/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("dtoken") != "token-value" || q.Get("platform") != "noun" || q.Get("account") != "user-a" {
			t.Fatalf("unexpected query: %#v", q)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"msg": "查询成功",
			"data": []map[string]interface{}{
				{
					"id":             "K1",
					"name":           "课程K",
					"kcjs":           "教师",
					"studyStartTime": "2026-03-01",
					"studyEndTime":   "2026-03-31",
				},
			},
		})
	}))
	defer srv.Close()

	result, err := kunCallQuery(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "noun", "school", "user-a", "pwd")
	if err != nil {
		t.Fatalf("kunCallQuery returned error: %v", err)
	}
	if result.Msg != "查询成功" || len(result.Data) != 1 || result.Data[0].ID != "K1" {
		t.Fatalf("unexpected kun query result: %+v", result)
	}
}

func TestYYYCallOrderAndProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"accessToken":  "ACCESS",
					"refreshToken": "REFRESH",
				},
				"message": "success",
			})
		case "/api/order":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    200,
				"message": "ok",
				"data":    []string{"YID-9"},
			})
		case "/api/getorder":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"list": []map[string]interface{}{
						{
							"id":     11,
							"odname": "user-a",
							"status": "原状态",
							"train":  "课程A",
							"code":   102,
							"note":   "完成",
						},
					},
				},
				"message": "success",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	sup := &model.SupplierFull{URL: srv.URL, User: "admin", Pass: "pwd", PT: "yyy"}
	orderResult, err := yyyCallOrder(sup, "user", "pass", "课程A", "site-1")
	if err != nil {
		t.Fatalf("yyyCallOrder returned error: %v", err)
	}
	if orderResult.Code != 1 || orderResult.YID != "YID-9" {
		t.Fatalf("unexpected yyy order result: %+v", orderResult)
	}

	progress, err := yyyQueryProgress(sup, "user-a", progressDebugInfo{})
	if err != nil {
		t.Fatalf("yyyQueryProgress returned error: %v", err)
	}
	if len(progress) != 1 || progress[0].Status != "已完成" || progress[0].Process != "100%" {
		t.Fatalf("unexpected yyy progress: %+v", progress)
	}
}

func TestYYYQueryBalanceReloginRetry(t *testing.T) {
	yyySessionsMu.Lock()
	yyySessions = map[string]*yyySession{}
	yyySessionsMu.Unlock()

	loginCalls := 0
	moneyCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			loginCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"accessToken":  "ACCESS",
					"refreshToken": "REFRESH",
				},
				"message": "success",
			})
		case "/api/money":
			moneyCalls++
			if moneyCalls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    302,
					"message": "请重新登录",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"money": "9.9",
				},
				"message": "success",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	result, err := yyyQueryBalance(&model.SupplierFull{URL: srv.URL, User: "admin", Pass: "pwd", PT: "yyy", HID: 1, Name: "YYY"})
	if err != nil {
		t.Fatalf("yyyQueryBalance returned error: %v", err)
	}
	if result["money"] != "9.9" {
		t.Fatalf("unexpected yyy money result: %+v", result)
	}
	if loginCalls != 2 || moneyCalls != 2 {
		t.Fatalf("expected relogin retry, got loginCalls=%d moneyCalls=%d", loginCalls, moneyCalls)
	}
}

func TestYYYGetClassesParsesPriceUnitFallback(t *testing.T) {
	yyySessionsMu.Lock()
	yyySessions = map[string]*yyySession{}
	yyySessionsMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"accessToken":  "ACCESS",
					"refreshToken": "REFRESH",
				},
				"message": "success",
			})
		case "/api/site":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"list": []map[string]interface{}{
						{
							"id":         1,
							"name":       "商品A",
							"trans":      "说明A",
							"price":      0,
							"price_unit": "1.5 /门",
						},
						{
							"id":    2,
							"name":  "商品B",
							"trans": "说明B",
							"price": 2.25,
						},
					},
				},
				"message": "success",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	items, err := yyyGetClasses(&model.SupplierFull{URL: srv.URL, User: "admin", Pass: "pwd", PT: "yyy", Name: "YYY"})
	if err != nil {
		t.Fatalf("yyyGetClasses returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item length: %+v", items)
	}
	if items[0].Price != 1.5 || items[0].CID != "1" || items[0].CategoryName != "YYY" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].Price != 2.25 || items[1].CID != "2" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}

func TestXxtCallQueryPhoneLoginFlow(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldPhone := xxtPhoneLoginURLTpl
	oldSchool := xxtSchoolSearchURLTpl
	oldID := xxtIDLoginURLTpl
	oldCourse := xxtCourseURL
	defer func() {
		xxtPhoneLoginURLTpl = oldPhone
		xxtSchoolSearchURLTpl = oldSchool
		xxtIDLoginURLTpl = oldID
		xxtCourseURL = oldCourse
	}()

	xxtPhoneLoginURLTpl = srv.URL + "/phone-login"
	xxtSchoolSearchURLTpl = srv.URL + "/school?filter=%s"
	xxtIDLoginURLTpl = srv.URL + "/id-login?fid=%s&idNumber=%s"
	xxtCourseURL = srv.URL + "/mycourse"

	mux.HandleFunc("/phone-login", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("uname") != "13800000000" || r.Form.Get("code") != "pwd" {
			t.Fatalf("unexpected phone login form: %#v", r.Form)
		}
		http.SetCookie(w, &http.Cookie{Name: "UID", Value: "u1"})
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": true})
	})
	mux.HandleFunc("/mycourse", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Cookie"), "UID=u1") {
			t.Fatalf("unexpected cookie header: %q", r.Header.Get("Cookie"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": 1,
			"channelList": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"course": map[string]interface{}{
							"data": []map[string]interface{}{
								{"id": "c1", "name": "课程一", "imageurl": "img.png"},
							},
						},
					},
				},
			},
		})
	})

	result, err := xxtCallQuery("13800000000", "pwd", "学校")
	if err != nil {
		t.Fatalf("xxtCallQuery returned error: %v", err)
	}
	if result["code"] != 1 || result["msg"] != "查询成功" {
		t.Fatalf("unexpected xxt result: %+v", result)
	}
	data, ok := result["data"].([]map[string]interface{})
	if !ok || len(data) != 1 || data[0]["id"] != "c1" {
		t.Fatalf("unexpected xxt data: %#v", result["data"])
	}
}

func TestXxtCallQuerySchoolLoginFlow(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldPhone := xxtPhoneLoginURLTpl
	oldSchool := xxtSchoolSearchURLTpl
	oldID := xxtIDLoginURLTpl
	oldCourse := xxtCourseURL
	defer func() {
		xxtPhoneLoginURLTpl = oldPhone
		xxtSchoolSearchURLTpl = oldSchool
		xxtIDLoginURLTpl = oldID
		xxtCourseURL = oldCourse
	}()

	xxtPhoneLoginURLTpl = srv.URL + "/phone-login"
	xxtSchoolSearchURLTpl = srv.URL + "/school?filter=%s"
	xxtIDLoginURLTpl = srv.URL + "/id-login?fid=%s&idNumber=%s"
	xxtCourseURL = srv.URL + "/mycourse"

	mux.HandleFunc("/school", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filter") != "学校A" {
			t.Fatalf("unexpected school query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"froms": []map[string]interface{}{
				{"schoolid": "FID-1"},
			},
		})
	})
	mux.HandleFunc("/id-login", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fid") != "FID-1" || r.URL.Query().Get("idNumber") != "20250001" {
			t.Fatalf("unexpected id login query: %#v", r.URL.Query())
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("pwd") != "pwd" {
			t.Fatalf("unexpected password form: %#v", r.Form)
		}
		http.SetCookie(w, &http.Cookie{Name: "UID", Value: "u2"})
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": true})
	})
	mux.HandleFunc("/mycourse", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Cookie"), "UID=u2") {
			t.Fatalf("unexpected cookie header: %q", r.Header.Get("Cookie"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": 1,
			"channelList": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"course": map[string]interface{}{
							"data": []map[string]interface{}{
								{"id": "c2", "name": "课程二", "imageurl": "img2.png"},
							},
						},
					},
				},
			},
		})
	})

	result, err := xxtCallQuery("20250001", "pwd", "学校A")
	if err != nil {
		t.Fatalf("xxtCallQuery school flow returned error: %v", err)
	}
	if result["code"] != 1 || result["userinfo"] != "学校A 20250001 pwd" {
		t.Fatalf("unexpected xxt school result: %+v", result)
	}
	data, ok := result["data"].([]map[string]interface{})
	if !ok || len(data) != 1 || data[0]["id"] != "c2" {
		t.Fatalf("unexpected xxt school data: %#v", result["data"])
	}
}

func TestXxtCallQuerySchoolNotFound(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldPhone := xxtPhoneLoginURLTpl
	oldSchool := xxtSchoolSearchURLTpl
	oldID := xxtIDLoginURLTpl
	oldCourse := xxtCourseURL
	defer func() {
		xxtPhoneLoginURLTpl = oldPhone
		xxtSchoolSearchURLTpl = oldSchool
		xxtIDLoginURLTpl = oldID
		xxtCourseURL = oldCourse
	}()

	xxtPhoneLoginURLTpl = srv.URL + "/phone-login"
	xxtSchoolSearchURLTpl = srv.URL + "/school?filter=%s"
	xxtIDLoginURLTpl = srv.URL + "/id-login?fid=%s&idNumber=%s"
	xxtCourseURL = srv.URL + "/mycourse"

	mux.HandleFunc("/school", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"froms": []map[string]interface{}{},
		})
	})

	result, err := xxtCallQuery("20250001", "pwd", "学校A")
	if err != nil {
		t.Fatalf("xxtCallQuery school-not-found returned error: %v", err)
	}
	if result["code"] != -1 || result["msg"] != "未找到学校信息" {
		t.Fatalf("unexpected school-not-found result: %+v", result)
	}
}

func TestXxtCallQueryPhoneLoginFailure(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldPhone := xxtPhoneLoginURLTpl
	oldSchool := xxtSchoolSearchURLTpl
	oldID := xxtIDLoginURLTpl
	oldCourse := xxtCourseURL
	defer func() {
		xxtPhoneLoginURLTpl = oldPhone
		xxtSchoolSearchURLTpl = oldSchool
		xxtIDLoginURLTpl = oldID
		xxtCourseURL = oldCourse
	}()

	xxtPhoneLoginURLTpl = srv.URL + "/phone-login"
	xxtSchoolSearchURLTpl = srv.URL + "/school?filter=%s"
	xxtIDLoginURLTpl = srv.URL + "/id-login?fid=%s&idNumber=%s"
	xxtCourseURL = srv.URL + "/mycourse"

	mux.HandleFunc("/phone-login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": false,
		})
	})

	result, err := xxtCallQuery("13800000000", "pwd", "学校A")
	if err != nil {
		t.Fatalf("xxtCallQuery phone-login-failure returned error: %v", err)
	}
	if result["code"] != -1 || result["msg"] != "信息错误或者重试" {
		t.Fatalf("unexpected phone-login-failure result: %+v", result)
	}
}

func TestQueryLonglongLogsParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/streamLogs" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "OID-1" || r.URL.Query().Get("key") != "secret" {
			t.Fatalf("unexpected query: %#v", r.URL.Query())
		}
		_, _ = w.Write([]byte("data: {\"time\":\"2026-03-12 10:00:00\",\"course\":\"课程A\",\"status\":\"进行中\",\"process\":\"50%\",\"message\":\"继续学习\"}\n\ndata: 纯文本日志\n\ndata: [DONE]\n"))
	}))
	defer srv.Close()

	logs, err := queryLonglongLogs(&model.SupplierFull{URL: srv.URL, Pass: "secret", PT: "longlong"}, "OID-1")
	if err != nil {
		t.Fatalf("queryLonglongLogs returned error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("unexpected logs length: %+v", logs)
	}
	if logs[0].Course != "课程A" || logs[0].Remarks != "继续学习" || logs[0].Process != "50%" {
		t.Fatalf("unexpected first log: %+v", logs[0])
	}
	if logs[1].Remarks != "纯文本日志" {
		t.Fatalf("unexpected second log: %+v", logs[1])
	}
}

func TestSimpleQueryProgressMapsStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Api/Query" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("token") != "token-value" || r.Form.Get("cid") != "noun" {
			t.Fatalf("unexpected form: %#v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"data": map[string]interface{}{
				"id":       "OID-1",
				"course":   "课程S",
				"user":     "user-s",
				"status":   "待处理",
				"progress": "30%",
				"process":  "",
				"kcks":     "2026-03-01",
				"kcjs":     "2026-03-31",
			},
		})
	}))
	defer srv.Close()

	progress, err := simpleQueryProgress(&model.SupplierFull{URL: srv.URL, Token: "token-value"}, "noun", "school", "user-s", "pwd", "课程S", "KC-1", progressDebugInfo{})
	if err != nil {
		t.Fatalf("simpleQueryProgress returned error: %v", err)
	}
	if len(progress) != 1 {
		t.Fatalf("unexpected progress length: %+v", progress)
	}
	if progress[0].Status != "进行中" || progress[0].Remarks != "暂无详情" || progress[0].Process != "30%" {
		t.Fatalf("unexpected simple progress item: %+v", progress[0])
	}
}

func TestQueryOrderProgressUsesConfiguredProgressParamNames(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["spi"] = PlatformConfig{
		AuthType:         "uid_key",
		SuccessCode:      "0",
		ReturnsYID:       true,
		ProgressPath:     "/api/search",
		ProgressMethod:   "POST",
		ProgressParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("uid") != "uid-value" || r.Form.Get("key") != "key-value" {
			t.Fatalf("unexpected auth form: %#v", r.Form)
		}
		if r.Form.Get("username") != "user-a" || r.Form.Get("kname") != "课程A" || r.Form.Get("cid") != "UP-COURSE-1" || r.Form.Get("yid") != "Y-1" {
			t.Fatalf("unexpected progress form: %#v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": []map[string]interface{}{
				{"id": "Y-1", "kcname": "课程A", "user": "user-a", "status": "进行中", "process": "50%", "remarks": "ok"},
			},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	progress, err := svc.QueryOrderProgress(&model.SupplierFull{PT: "spi", URL: srv.URL, User: "uid-value", Pass: "key-value"}, "Y-1", "user-a", map[string]string{
		"kcname": "课程A",
		"noun":   "UP-COURSE-1",
		"kcid":   "CID-1",
	})
	if err != nil {
		t.Fatalf("QueryOrderProgress returned error: %v", err)
	}
	if len(progress) != 1 || progress[0].YID != "Y-1" {
		t.Fatalf("unexpected progress result: %+v", progress)
	}
}

func TestQueryOrderProgressWithoutYIDStillUsesSingleConfiguredEndpoint(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["single-progress"] = PlatformConfig{
		AuthType:         "uid_key",
		SuccessCode:      "0",
		ProgressPath:     "/api.php?act=chadan2",
		ProgressMethod:   "POST",
		ProgressParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api.php" || r.URL.Query().Get("act") != "chadan2" {
			t.Fatalf("unexpected endpoint: %s", r.URL.String())
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("uid") != "uid-value" || r.Form.Get("key") != "key-value" || r.Form.Get("username") != "user-a" {
			t.Fatalf("unexpected progress form: %#v", r.Form)
		}
		if got := r.Form.Get("yid"); got != "" {
			t.Fatalf("expected empty yid when not provided, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": []map[string]interface{}{
				{"id": "OID-1", "kcname": "课程A", "user": "user-a", "status": "进行中", "process": "50%", "remarks": "ok"},
			},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	progress, err := svc.QueryOrderProgress(&model.SupplierFull{PT: "single-progress", URL: srv.URL, User: "uid-value", Pass: "key-value"}, "", "user-a", map[string]string{
		"kcname": "课程A",
		"noun":   "UP-COURSE-1",
	})
	if err != nil {
		t.Fatalf("QueryOrderProgress returned error: %v", err)
	}
	if len(progress) != 1 || progress[0].User != "user-a" {
		t.Fatalf("unexpected progress result: %+v", progress)
	}
}

func TestQueryBatchOrderProgressUsesConfiguredEndpoint(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["batch-progress"] = PlatformConfig{
		AuthType:              "uid_key",
		SuccessCode:           "0",
		BatchProgressPath:     "/api.php?act=pljd",
		BatchProgressMethod:   "POST",
		BatchProgressParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","count":"{{batch.count}}","yids":"{{batch.yids}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api.php" || r.URL.Query().Get("act") != "pljd" {
			t.Fatalf("unexpected endpoint: %s", r.URL.String())
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("uid") != "uid-value" || r.Form.Get("key") != "key-value" {
			t.Fatalf("unexpected auth form: %#v", r.Form)
		}
		if r.Form.Get("count") != "2" || r.Form.Get("yids") != "Y-1,Y-2" {
			t.Fatalf("unexpected batch form: %#v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": []map[string]interface{}{
				{"id": "Y-1", "cid": "NOUN-1", "status": "进行中", "process": "50%", "remarks": "ok", "ptname": "不应作为课程名"},
				{"id": "Y-2", "kcname": "课程B", "user": "user-b", "status_text": "已完成", "process": "100%"},
			},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	progress, err := svc.QueryBatchOrderProgress(&model.SupplierFull{PT: "batch-progress", URL: srv.URL, User: "uid-value", Pass: "key-value"}, []model.SupplierBatchProgressRef{
		{YID: "Y-1", User: "user-a", KCName: "课程A", KCID: "CID-1", Noun: "NOUN-1"},
		{YID: "Y-2", User: "user-b", KCName: "课程B", KCID: "CID-2", Noun: "NOUN-2"},
	})
	if err != nil {
		t.Fatalf("QueryBatchOrderProgress returned error: %v", err)
	}
	if len(progress) != 2 || progress[0].YID != "Y-1" || progress[0].Noun != "NOUN-1" || progress[0].KCName != "" || progress[1].StatusText != "已完成" {
		t.Fatalf("unexpected batch progress result: %+v", progress)
	}
}

func TestQueryBatchOrderProgressTreatsNoUpdateMessageAsEmptySuccess(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["batch-progress-empty"] = PlatformConfig{
		AuthType:              "uid_key",
		SuccessCode:           "0",
		BatchProgressPath:     "/api.php?act=pljd",
		BatchProgressMethod:   "POST",
		BatchProgressParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": -1,
			"msg":  "该时间范围内无可更新进度订单",
			"data": []map[string]interface{}{},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	progress, err := svc.QueryBatchOrderProgress(&model.SupplierFull{PT: "batch-progress-empty", URL: srv.URL, User: "uid-value", Pass: "key-value"}, []model.SupplierBatchProgressRef{
		{YID: "Y-1", User: "user-a", KCName: "课程A", KCID: "CID-1", Noun: "NOUN-1"},
	})
	if err != nil {
		t.Fatalf("QueryBatchOrderProgress returned error: %v", err)
	}
	if len(progress) != 0 {
		t.Fatalf("expected empty batch progress result, got %+v", progress)
	}
}

func TestCallSupplierOrderUsesConfiguredParamMap(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-order"] = PlatformConfig{
		SuccessCode:   "0",
		ReturnsYID:    true,
		OrderPath:     "/api/order/create",
		OrderMethod:   http.MethodPost,
		OrderBodyType: "json",
		OrderParamMap: `{"token":"{{supplier.key}}","account":"{{order.user}}","password":"{{order.pass}}","courseName":"{{order.kcname}}","courseId":"{{order.kcid}}","platformId":"{{order.noun}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/order/create" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("unexpected content-type: %s", ct)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if body["token"] != "key-1" || body["account"] != "user-1" || body["password"] != "pwd-1" ||
			body["courseName"] != "课程A" || body["courseId"] != "KC-11" || body["platformId"] != "UP-88" {
			t.Fatalf("unexpected order payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "id": "Y-998", "msg": "ok"})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	result, err := svc.CallSupplierOrder(
		&model.SupplierFull{PT: "mapped-order", URL: srv.URL, User: "uid-1", Pass: "key-1"},
		&model.ClassFull{Noun: "UP-88"},
		"学校A",
		"user-1",
		"pwd-1",
		"KC-11",
		"课程A",
		nil,
	)
	if err != nil {
		t.Fatalf("CallSupplierOrder returned error: %v", err)
	}
	if result.YID != "Y-998" || result.Code != 1 {
		t.Fatalf("unexpected order result: %+v", result)
	}
}

func TestChangePasswordUsesActionNewPasswordTemplateVar(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-change-pass"] = PlatformConfig{
		ChangePassPath:     "/api/order/change-password",
		ChangePassMethod:   http.MethodPost,
		ChangePassBodyType: "json",
		ChangePassParamMap: `{"token":"{{supplier.key}}","orderId":"{{order.yid}}","newPassword":"{{action.new_password}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/order/change-password" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("unexpected content-type: %s", ct)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if body["token"] != "key-1" || body["orderId"] != "OID-1" || body["newPassword"] != "fresh-pass" {
			t.Fatalf("unexpected change password payload: %#v", body)
		}
		if _, exists := body["newPwd"]; exists {
			t.Fatalf("unexpected legacy newPwd payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "msg": "ok"})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	code, msg, err := svc.ChangePassword(
		&model.SupplierFull{PT: "mapped-change-pass", URL: srv.URL, User: "uid-1", Pass: "key-1"},
		"OID-1",
		"fresh-pass",
	)
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	if code != 0 || msg != "ok" {
		t.Fatalf("unexpected change password result: code=%d msg=%q", code, msg)
	}
}

func TestCallSupplierQueryUsesActionTemplateVars(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-query"] = PlatformConfig{
		QueryPath:     "/api/query/course",
		QueryMethod:   http.MethodPost,
		QueryBodyType: "json",
		QueryParamMap: `{"token":"{{supplier.key}}","school":"{{action.school}}","account":"{{action.user}}","password":"{{action.password}}","platform":"{{action.platform}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/query/course" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if body["token"] != "key-1" || body["school"] != "学校A" || body["account"] != "user-1" ||
			body["password"] != "pwd-1" || body["platform"] != "UP-88" {
			t.Fatalf("unexpected query payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":     0,
			"msg":      "查询成功",
			"userName": "张三",
			"data": []map[string]interface{}{
				{"id": "K1", "name": "课程A"},
			},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	result, err := svc.callSupplierQuery(
		&model.SupplierFull{PT: "mapped-query", URL: srv.URL, User: "uid-1", Pass: "key-1"},
		&model.ClassFull{Noun: "UP-88"},
		"学校A",
		"user-1",
		"pwd-1",
	)
	if err != nil {
		t.Fatalf("callSupplierQuery returned error: %v", err)
	}
	if result.UserName != "张三" || len(result.Data) != 1 || result.Data[0].ID != "K1" {
		t.Fatalf("unexpected query result: %+v", result)
	}
}

func TestCallSupplierQueryRequiresExplicitQueryConfig(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["missing-query-config"] = PlatformConfig{}
	dbConfigMu.Unlock()

	svc := &Service{}
	_, err := svc.callSupplierQuery(
		&model.SupplierFull{PT: "missing-query-config"},
		&model.ClassFull{Noun: "UP-1"},
		"学校A",
		"user-1",
		"pwd-1",
	)
	if err == nil {
		t.Fatal("expected missing query config error")
	}
	if !strings.Contains(err.Error(), "查课接口未配置路径") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildActionParamsDoesNotExposeLegacyOrderNewPwdVar(t *testing.T) {
	params, err := buildActionParams(
		`{"legacy":"{{order.new_pwd}}","current":"{{action.new_password}}"}`,
		&model.SupplierFull{},
		map[string]string{"action.new_password": "fresh-pass"},
		nil,
	)
	if err != nil {
		t.Fatalf("buildActionParams returned error: %v", err)
	}
	if params["current"] != "fresh-pass" {
		t.Fatalf("expected action.new_password to resolve, got %#v", params)
	}
	if _, exists := params["legacy"]; exists {
		t.Fatalf("expected legacy order.new_pwd to be unavailable, got %#v", params)
	}
}

func TestBuildActionParamsIgnoresUnnamespacedTemplateVars(t *testing.T) {
	params, err := buildActionParams(
		`{"scoped":"{{action.user}}","raw":"{{user}}"}`,
		&model.SupplierFull{},
		map[string]string{"action.user": "user-a", "user": "user-a"},
		nil,
	)
	if err != nil {
		t.Fatalf("buildActionParams returned error: %v", err)
	}
	if params["scoped"] != "user-a" {
		t.Fatalf("expected scoped action var to resolve, got %#v", params)
	}
	if _, exists := params["raw"]; exists {
		t.Fatalf("expected unnamespaced raw var to be unavailable, got %#v", params)
	}
}

func TestSubmitAndQueryReportUseActionTemplateVars(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-report"] = PlatformConfig{
		ReportPath:        "/api/report/submit",
		ReportMethod:      http.MethodPost,
		ReportBodyType:    "json",
		ReportParamMap:    `{"token":"{{supplier.key}}","orderId":"{{order.yid}}","type":"{{action.ticket_type}}","content":"{{action.content}}"}`,
		GetReportPath:     "/api/report/query",
		GetReportMethod:   http.MethodPost,
		GetReportBodyType: "json",
		GetReportParamMap: `{"token":"{{supplier.key}}","reportId":"{{action.report_id}}"}`,
	}
	dbConfigMu.Unlock()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/api/report/submit", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode submit body failed: %v", err)
		}
		if body["token"] != "key-1" || body["orderId"] != "OID-1" || body["type"] != "after-sale" || body["content"] != "content-1" {
			t.Fatalf("unexpected submit payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 1, "msg": "submitted", "workId": 321})
	})
	mux.HandleFunc("/api/report/query", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode query body failed: %v", err)
		}
		if body["token"] != "key-1" || body["reportId"] != "321" {
			t.Fatalf("unexpected query payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"data": map[string]interface{}{"answer": "done", "state": "closed"},
		})
	})

	svc := &Service{client: srv.Client()}
	sup := &model.SupplierFull{PT: "mapped-report", URL: srv.URL, User: "uid-1", Pass: "key-1"}

	code, workID, msg, err := svc.SubmitReport(sup, "OID-1", "after-sale", "content-1")
	if err != nil {
		t.Fatalf("SubmitReport returned error: %v", err)
	}
	if code != 1 || workID != 321 || msg != "submitted" {
		t.Fatalf("unexpected submit result: code=%d workID=%d msg=%q", code, workID, msg)
	}

	code, answer, state, err := svc.QueryReport(sup, "321")
	if err != nil {
		t.Fatalf("QueryReport returned error: %v", err)
	}
	if code != 1 || answer != "done" || state != "closed" {
		t.Fatalf("unexpected query result: code=%d answer=%q state=%q", code, answer, state)
	}
}

func TestQueryOrderLogsUsesConfiguredParamMap(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-log"] = PlatformConfig{
		LogPath:     "/logs/search",
		LogMethod:   http.MethodGet,
		LogBodyType: "query",
		LogParamMap: `{"token":"{{supplier.token}}","account":"{{order.user}}","password":"{{order.pass}}","course":"{{order.kcname}}","courseId":"{{order.kcid}}"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logs/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("token") != "token-1" || q.Get("account") != "user-log" || q.Get("password") != "pass-log" ||
			q.Get("course") != "课程Log" || q.Get("courseId") != "CID-LOG" {
			t.Fatalf("unexpected log query: %#v", q)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "logs": []string{"line-1"}})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	logs, err := svc.QueryOrderLogs(
		&model.SupplierFull{PT: "mapped-log", URL: srv.URL, Token: "token-1"},
		"OID-1",
		map[string]string{"user": "user-log", "pass": "pass-log", "kcname": "课程Log", "kcid": "CID-LOG"},
	)
	if err != nil {
		t.Fatalf("QueryOrderLogs returned error: %v", err)
	}
	if len(logs) == 0 || logs[len(logs)-1].Remarks != "line-1" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestGetSupplierCategoriesUsesConfiguredParamMap(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-catalog"] = PlatformConfig{
		AuthType:         "uid_key",
		CategoryPath:     "/api/catalog/categories",
		CategoryMethod:   http.MethodPost,
		CategoryBodyType: "json",
		CategoryParamMap: `{"token":"{{supplier.key}}","scope":"catalog"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/catalog/categories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("unexpected content-type: %s", ct)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		if body["token"] != "key-1" || body["scope"] != "catalog" {
			t.Fatalf("unexpected category payload: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": []map[string]interface{}{
				{"fid": "10", "fname": "分类A"},
			},
		})
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	cats := svc.GetSupplierCategories(&model.SupplierFull{PT: "mapped-catalog", URL: srv.URL, User: "uid-1", Pass: "key-1"})
	if cats["10"] != "分类A" {
		t.Fatalf("unexpected categories: %#v", cats)
	}
}

func TestGetSupplierClassesUsesConfiguredParamMap(t *testing.T) {
	withRegistryPlatformConfigs(t)
	dbConfigMu.Lock()
	dbConfigCache["mapped-class-list"] = PlatformConfig{
		AuthType:          "uid_key",
		ClassListPath:     "/api/catalog/classes",
		ClassListMethod:   http.MethodGet,
		ClassListBodyType: "query",
		ClassListParamMap: `{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","catalog":"classes"}`,
	}
	dbConfigMu.Unlock()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/catalog/classes":
			q := r.URL.Query()
			if q.Get("uid") != "uid-1" || q.Get("key") != "key-1" || q.Get("catalog") != "classes" {
				t.Fatalf("unexpected class list query: %#v", q)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []map[string]interface{}{
					{"cid": "C-1", "name": "课程A", "price": "12.5", "fenlei": "10", "content": "desc", "category_name": "分类A"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	svc := &Service{client: srv.Client()}
	items, err := svc.GetSupplierClasses(&model.SupplierFull{PT: "mapped-class-list", URL: srv.URL, User: "uid-1", Pass: "key-1"})
	if err != nil {
		t.Fatalf("GetSupplierClasses returned error: %v", err)
	}
	if len(items) != 1 || items[0].CID != "C-1" || items[0].CategoryName != "分类A" {
		t.Fatalf("unexpected class list: %+v", items)
	}
}

func TestSubmitAndQueryReport(t *testing.T) {
	withRegistryPlatformConfigs(t)

	t.Run("StandardFormStyle", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		defer srv.Close()

		svc := &Service{client: srv.Client()}
		sup := &model.SupplierFull{
			PT:   "liunian",
			URL:  srv.URL,
			User: "uid-1",
			Pass: "key-1",
		}

		mux.HandleFunc("/api.php", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form failed: %v", err)
			}
			switch r.URL.Query().Get("act") {
			case "submitWorkOrder":
				if r.Form.Get("uid") != "uid-1" || r.Form.Get("key") != "key-1" ||
					r.Form.Get("id") != "OID-1" || r.Form.Get("question") != "content-1" {
					t.Fatalf("unexpected submit form: %#v", r.Form)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 1,
					"msg":  "submitted",
					"data": map[string]interface{}{"reportId": 321},
				})
			case "queryWorkOrder":
				if r.Form.Get("uid") != "uid-1" || r.Form.Get("key") != "key-1" ||
					r.Form.Get("reportId") != "321" {
					t.Fatalf("unexpected query form: %#v", r.Form)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 1,
					"data": map[string]interface{}{
						"answer": "done",
						"state":  "closed",
					},
				})
			default:
				t.Fatalf("unexpected act: %s", r.URL.RawQuery)
			}
		})

		code, workID, msg, err := svc.SubmitReport(sup, "OID-1", "", "content-1")
		if err != nil {
			t.Fatalf("SubmitReport standard returned error: %v", err)
		}
		if code != 1 || workID != 321 || msg != "submitted" {
			t.Fatalf("unexpected standard submit result: code=%d workID=%d msg=%q", code, workID, msg)
		}

		code, answer, state, err := svc.QueryReport(sup, "321")
		if err != nil {
			t.Fatalf("QueryReport standard returned error: %v", err)
		}
		if code != 1 || answer != "done" || state != "closed" {
			t.Fatalf("unexpected standard query result: code=%d answer=%q state=%q", code, answer, state)
		}
	})

	t.Run("TokenStyle", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		defer srv.Close()

		svc := &Service{client: srv.Client()}
		sup := &model.SupplierFull{
			PT:    "2xx",
			URL:   srv.URL,
			Token: "token-2xx",
		}

		mux.HandleFunc("/api/submitWork", func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode submit body failed: %v", err)
			}
			if body["token"] != "token-2xx" || body["type"] != "after-sale" ||
				body["id"] != "OID-2" || body["content"] != "content-2" {
				t.Fatalf("unexpected submit body: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":   1,
				"msg":    "accepted",
				"workId": 654,
			})
		})

		mux.HandleFunc("/api/queryWork", func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode query body failed: %v", err)
			}
			if body["token"] != "token-2xx" || body["workId"] != "654" {
				t.Fatalf("unexpected query body: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 1,
				"data": map[string]interface{}{
					"answer": "processing",
					"status": 2,
				},
			})
		})

		code, workID, msg, err := svc.SubmitReport(sup, "OID-2", "after-sale", "content-2")
		if err != nil {
			t.Fatalf("SubmitReport token returned error: %v", err)
		}
		if code != 1 || workID != 654 || msg != "accepted" {
			t.Fatalf("unexpected token submit result: code=%d workID=%d msg=%q", code, workID, msg)
		}

		code, answer, state, err := svc.QueryReport(sup, "654")
		if err != nil {
			t.Fatalf("QueryReport token returned error: %v", err)
		}
		if code != 1 || answer != "processing" || state != "2" {
			t.Fatalf("unexpected token query result: code=%d answer=%q state=%q", code, answer, state)
		}
	})
}

func TestSubmitAndQueryReportFailures(t *testing.T) {
	withRegistryPlatformConfigs(t)

	t.Run("SubmitReportInvalidJSON", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		defer srv.Close()

		svc := &Service{client: srv.Client()}
		sup := &model.SupplierFull{
			PT:   "liunian",
			URL:  srv.URL,
			User: "uid-1",
			Pass: "key-1",
		}

		mux.HandleFunc("/api.php", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("act") != "submitWorkOrder" {
				t.Fatalf("unexpected act: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte("not-json"))
		})

		code, workID, msg, err := svc.SubmitReport(sup, "OID-3", "", "content-3")
		if err == nil || !strings.Contains(err.Error(), "上游返回解析失败") {
			t.Fatalf("expected parse error, got code=%d workID=%d msg=%q err=%v", code, workID, msg, err)
		}
	})

	t.Run("QueryReportInvalidJSON", func(t *testing.T) {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		defer srv.Close()

		svc := &Service{client: srv.Client()}
		sup := &model.SupplierFull{
			PT:    "2xx",
			URL:   srv.URL,
			Token: "token-2xx",
		}

		mux.HandleFunc("/api/queryWork", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("bad-response"))
		})

		code, answer, state, err := svc.QueryReport(sup, "654")
		if err == nil || !strings.Contains(err.Error(), "上游返回解析失败") {
			t.Fatalf("expected parse error, got code=%d answer=%q state=%q err=%v", code, answer, state, err)
		}
	})

	t.Run("SubmitReportRequestFailure", func(t *testing.T) {
		srv := httptest.NewServer(http.NewServeMux())
		client := srv.Client()
		url := srv.URL
		srv.Close()

		svc := &Service{client: client}
		sup := &model.SupplierFull{
			PT:   "liunian",
			URL:  url,
			User: "uid-1",
			Pass: "key-1",
		}

		code, workID, msg, err := svc.SubmitReport(sup, "OID-4", "", "content-4")
		if err == nil || !strings.Contains(err.Error(), "请求上游失败") {
			t.Fatalf("expected request error, got code=%d workID=%d msg=%q err=%v", code, workID, msg, err)
		}
	})
}
