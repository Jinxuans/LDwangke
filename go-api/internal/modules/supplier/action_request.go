package supplier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go-api/internal/model"
)

type actionExecutionResult struct {
	Body        []byte
	Status      string
	Method      string
	URL         string
	ContentType string
	Payload     string
}

var actionTemplatePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

func buildActionTemplateVars(sup *model.SupplierFull, fields map[string]string) map[string]string {
	vars := map[string]string{
		"supplier.uid":    sup.User,
		"supplier.key":    sup.Pass,
		"supplier.pass":   sup.Pass,
		"supplier.token":  getSupplierToken(sup),
		"supplier.cookie": sup.Cookie,
		"supplier.url":    sup.URL,
		"supplier.pt":     sup.PT,
	}

	for k, v := range fields {
		if k == "" {
			continue
		}
		vars[k] = v
		if !strings.Contains(k, ".") {
			vars["order."+k] = v
			vars["input."+k] = v
		}
	}

	if v := fields["user"]; v != "" {
		vars["username"] = v
		vars["order.username"] = v
	}
	if v := fields["pass"]; v != "" {
		vars["password"] = v
		vars["order.password"] = v
	}
	if v := fields["kcname"]; v != "" {
		vars["course_name"] = v
	}
	if v := fields["kcid"]; v != "" {
		vars["course_id"] = v
	}
	if v := fields["noun"]; v != "" {
		vars["platform_id"] = v
	}

	return vars
}

func renderActionTemplate(raw string, vars map[string]string) string {
	if raw == "" {
		return ""
	}
	return actionTemplatePattern.ReplaceAllStringFunc(raw, func(match string) string {
		submatch := actionTemplatePattern.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return ""
		}
		key := strings.TrimSpace(submatch[1])
		if val, ok := vars[key]; ok {
			return val
		}
		return ""
	})
}

func buildActionParams(paramMapRaw string, sup *model.SupplierFull, fields map[string]string, defaults map[string]string) (map[string]string, error) {
	if strings.TrimSpace(paramMapRaw) == "" {
		result := map[string]string{}
		for k, v := range defaults {
			if strings.TrimSpace(v) != "" {
				result[k] = v
			}
		}
		return result, nil
	}

	templateMap := map[string]string{}
	if err := json.Unmarshal([]byte(paramMapRaw), &templateMap); err != nil {
		return nil, fmt.Errorf("参数映射 JSON 格式错误: %v", err)
	}

	vars := buildActionTemplateVars(sup, fields)
	result := map[string]string{}
	for k, tmpl := range templateMap {
		if strings.TrimSpace(k) == "" {
			continue
		}
		value := renderActionTemplate(tmpl, vars)
		if strings.TrimSpace(value) == "" {
			continue
		}
		result[k] = value
	}
	return result, nil
}

func normalizeActionMethod(method, fallback string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = strings.ToUpper(strings.TrimSpace(fallback))
	}
	if method == "" {
		method = http.MethodPost
	}
	return method
}

func normalizeActionBodyType(bodyType, fallback, method string) string {
	bodyType = strings.ToLower(strings.TrimSpace(bodyType))
	if bodyType == "" {
		bodyType = strings.ToLower(strings.TrimSpace(fallback))
	}
	if bodyType == "" {
		if method == http.MethodGet {
			return "query"
		}
		return "form"
	}
	return bodyType
}

func normalizeSupplierBaseURL(rawURL string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(rawURL), "/")
	if baseURL != "" && !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func resolveConfiguredActionURL(rawURL, path string) string {
	baseURL := normalizeSupplierBaseURL(rawURL)
	path = strings.TrimSpace(path)
	if path != "" {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return path
		}
		if strings.HasPrefix(path, "/") {
			return baseURL + path
		}
		if baseURL == "" {
			return path
		}
		return baseURL + "/" + path
	}
	return baseURL
}

func getSupplierAuthToken(sup *model.SupplierFull, authType string) string {
	if strings.EqualFold(strings.TrimSpace(authType), "token_field") {
		return sup.Token
	}
	return getSupplierToken(sup)
}

func defaultSupplierAuthParams(sup *model.SupplierFull, authType string) map[string]string {
	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "", "uid_key":
		return map[string]string{
			"uid": sup.User,
			"key": sup.Pass,
		}
	case "token_only", "token_field":
		token := getSupplierAuthToken(sup, authType)
		if strings.TrimSpace(token) == "" {
			return map[string]string{}
		}
		return map[string]string{"token": token}
	case "none":
		return map[string]string{}
	default:
		return map[string]string{
			"uid": sup.User,
			"key": sup.Pass,
		}
	}
}

func prepareActionRequest(apiURL, method, bodyType string, params map[string]string) (*http.Request, string, string, error) {
	method = normalizeActionMethod(method, http.MethodPost)
	bodyType = normalizeActionBodyType(bodyType, "", method)

	values := url.Values{}
	for k, v := range params {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		values.Set(k, v)
	}

	switch {
	case method == http.MethodGet || bodyType == "query":
		if len(values) > 0 {
			if strings.Contains(apiURL, "?") {
				apiURL += "&" + values.Encode()
			} else {
				apiURL += "?" + values.Encode()
			}
		}
		req, err := http.NewRequest(method, apiURL, nil)
		return req, "", "", err
	case bodyType == "json":
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, "", "", err
		}
		req, err := http.NewRequest(method, apiURL, bytes.NewReader(jsonData))
		if err != nil {
			return nil, "", "", err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, "application/json", string(jsonData), nil
	default:
		encoded := values.Encode()
		req, err := http.NewRequest(method, apiURL, strings.NewReader(encoded))
		if err != nil {
			return nil, "", "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, "application/x-www-form-urlencoded", encoded, nil
	}
}

func (s *Service) executeConfiguredAction(
	sup *model.SupplierFull,
	apiURL string,
	method string,
	bodyType string,
	paramMap string,
	fallbackMethod string,
	fallbackBodyType string,
	defaultParams map[string]string,
	fields map[string]string,
) (*actionExecutionResult, error) {
	return s.executeConfiguredActionWithClient(
		s.client,
		sup,
		apiURL,
		method,
		bodyType,
		paramMap,
		fallbackMethod,
		fallbackBodyType,
		defaultParams,
		fields,
	)
}

func (s *Service) executeConfiguredActionWithClient(
	client *http.Client,
	sup *model.SupplierFull,
	apiURL string,
	method string,
	bodyType string,
	paramMap string,
	fallbackMethod string,
	fallbackBodyType string,
	defaultParams map[string]string,
	fields map[string]string,
) (*actionExecutionResult, error) {
	params, err := buildActionParams(paramMap, sup, fields, defaultParams)
	if err != nil {
		return nil, err
	}

	req, contentType, payload, err := prepareActionRequest(
		apiURL,
		normalizeActionMethod(method, fallbackMethod),
		normalizeActionBodyType(bodyType, fallbackBodyType, normalizeActionMethod(method, fallbackMethod)),
		params,
	)
	if err != nil {
		return nil, err
	}

	waitSupplierHost(sup)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &actionExecutionResult{
		Body:        body,
		Status:      resp.Status,
		Method:      req.Method,
		URL:         req.URL.String(),
		ContentType: contentType,
		Payload:     payload,
	}, nil
}
