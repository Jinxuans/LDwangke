package supplier

import (
	"regexp"
	"strings"

	"go-api/internal/model"
)

type actionTemplateContext struct {
	vars map[string]string
}

var actionVariablePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

func newActionTemplateContext(sup *model.SupplierFull, fields map[string]string) actionTemplateContext {
	ctx := actionTemplateContext{vars: map[string]string{}}
	ctx.addSupplier(sup)
	ctx.addFields(fields)
	return ctx
}

func (ctx actionTemplateContext) addSupplier(sup *model.SupplierFull) {
	if sup == nil {
		return
	}
	ctx.vars["supplier.uid"] = sup.User
	ctx.vars["supplier.key"] = sup.Pass
	ctx.vars["supplier.pass"] = sup.Pass
	ctx.vars["supplier.token"] = getSupplierToken(sup)
	ctx.vars["supplier.cookie"] = sup.Cookie
	ctx.vars["supplier.url"] = sup.URL
	ctx.vars["supplier.pt"] = sup.PT
}

func (ctx actionTemplateContext) addFields(fields map[string]string) {
	for key, value := range fields {
		ctx.set(key, value)
	}
}

func (ctx actionTemplateContext) set(key string, value string) {
	key = strings.TrimSpace(key)
	if key == "" || !strings.Contains(key, ".") {
		return
	}
	ctx.vars[key] = value
}

func (ctx actionTemplateContext) render(raw string) (string, error) {
	value := actionVariablePattern.ReplaceAllStringFunc(raw, func(match string) string {
		submatch := actionVariablePattern.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return ""
		}
		key := strings.TrimSpace(submatch[1])
		if val, ok := ctx.vars[key]; ok {
			return val
		}
		return ""
	})
	return value, nil
}

func orderActionFields(fields map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range fields {
		if strings.TrimSpace(value) == "" {
			continue
		}
		result["order."+key] = value
	}
	return result
}

func queryActionFields(school, user, pass, platform string) map[string]string {
	return map[string]string{
		"action.school":   school,
		"action.user":     user,
		"action.password": pass,
		"action.platform": platform,
	}
}

func orderProgressActionFields(yid, username string, extra map[string]string) map[string]string {
	fields := map[string]string{}
	if username != "" {
		fields["user"] = username
	}
	if yid != "" {
		fields["yid"] = yid
	}
	for key, value := range extra {
		if strings.TrimSpace(value) == "" || strings.HasPrefix(key, "__") {
			continue
		}
		fields[key] = value
	}
	return orderActionFields(fields)
}

func orderLogActionFields(yid string, extra map[string]string) map[string]string {
	fields := map[string]string{}
	if yid != "" {
		fields["yid"] = yid
	}
	for key, value := range extra {
		if strings.TrimSpace(value) == "" {
			continue
		}
		fields[key] = value
	}
	return orderActionFields(fields)
}
