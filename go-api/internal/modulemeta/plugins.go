package modulemeta

type PluginSpec struct {
	Name             string
	Kind             string
	Status           string
	HasRoutes        bool
	HasPluginRuntime bool
	RuntimeOwner     string
	RetirementRule   string
	RuntimeEntry     string
	Notes            string
}

var CronBridgePlugins = []string{
	"sdxy",
	"w",
	"xm",
	"ydsj",
	"yongye",
}

var PluginSpecs = map[string]PluginSpec{
	"appui": {
		Name:           "appui",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "AppUI product plugin.",
	},
	"paper": {
		Name:           "paper",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "Paper generation plugin.",
	},
	"sdxy": {
		Name:             "sdxy",
		Kind:             "plugin",
		Status:           "active",
		HasRoutes:        true,
		HasPluginRuntime: true,
		RuntimeOwner:     "cron_bridge",
		RetirementRule:   "Move off cron_bridge only when SDXY cron has a dedicated pluginruntime owner or the plugin is retired.",
		RuntimeEntry:     "pluginruntime/cron_bridge.go:RunSDXYCron",
		Notes:            "SDXY plugin with background sync runtime; cron loop still bridged through pluginruntime/cron_bridge.go.",
	},
	"sxdk": {
		Name:           "sxdk",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "SXDK plugin.",
	},
	"tuboshu": {
		Name:           "tuboshu",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "Tuboshu writing plugin.",
	},
	"tutuqg": {
		Name:           "tutuqg",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "TutuQG plugin.",
	},
	"tuzhi": {
		Name:           "tuzhi",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "TuZhi plugin.",
	},
	"w": {
		Name:             "w",
		Kind:             "plugin",
		Status:           "active",
		HasRoutes:        true,
		HasPluginRuntime: true,
		RuntimeOwner:     "cron_bridge",
		RetirementRule:   "Move off cron_bridge only when W cron has a dedicated pluginruntime owner or the plugin is retired.",
		RuntimeEntry:     "pluginruntime/cron_bridge.go:RunWCron",
		Notes:            "W plugin with background sync runtime; cron loop still bridged through pluginruntime/cron_bridge.go.",
	},
	"xm": {
		Name:             "xm",
		Kind:             "plugin",
		Status:           "active",
		HasRoutes:        true,
		HasPluginRuntime: true,
		RuntimeOwner:     "cron_bridge",
		RetirementRule:   "Move off cron_bridge only when XM cron has a dedicated pluginruntime owner or the plugin is retired.",
		RuntimeEntry:     "pluginruntime/cron_bridge.go:RunXMCron",
		Notes:            "XM plugin with background sync runtime; cron loop still bridged through pluginruntime/cron_bridge.go.",
	},
	"ydsj": {
		Name:             "ydsj",
		Kind:             "plugin",
		Status:           "active",
		HasRoutes:        true,
		HasPluginRuntime: true,
		RuntimeOwner:     "cron_bridge",
		RetirementRule:   "Move off cron_bridge only when YDSJ cron has a dedicated pluginruntime owner or the plugin is retired.",
		RuntimeEntry:     "pluginruntime/cron_bridge.go:RunYDSJCron",
		Notes:            "YDSJ plugin with background sync runtime; cron loop still bridged through pluginruntime/cron_bridge.go.",
	},
	"yfdk": {
		Name:           "yfdk",
		Kind:           "plugin",
		Status:         "active",
		HasRoutes:      true,
		RuntimeOwner:   "module",
		RetirementRule: "Retire only after routes and product obligations are removed.",
		RuntimeEntry:   "",
		Notes:          "YFDK plugin.",
	},
	"yongye": {
		Name:             "yongye",
		Kind:             "plugin",
		Status:           "active",
		HasRoutes:        true,
		HasPluginRuntime: true,
		RuntimeOwner:     "cron_bridge",
		RetirementRule:   "Move off cron_bridge only when Yongye cron has a dedicated pluginruntime owner or the plugin is retired.",
		RuntimeEntry:     "pluginruntime/cron_bridge.go:RunYongyeCron",
		Notes:            "Yongye plugin with background sync runtime; cron loop still bridged through pluginruntime/cron_bridge.go.",
	},
}
