package modulemeta

var CoreModules = map[string]bool{
	"admin":     true,
	"agent":     true,
	"auth":      true,
	"auxiliary": true,
	"chat":      true,
	"checkin":   true,
	"class":     true,
	"mail":      true,
	"order":     true,
	"push":      true,
	"supplier":  true,
	"tenant":    true,
	"user":      true,
}

var PluginModules = map[string]bool{
	"appui":   true,
	"paper":   true,
	"sdxy":    true,
	"sxdk":    true,
	"tuboshu": true,
	"tutuqg":  true,
	"tuzhi":   true,
	"w":       true,
	"xm":      true,
	"ydsj":    true,
	"yfdk":    true,
	"yongye":  true,
}

var SharedModules = map[string]bool{
	"common": true,
}
