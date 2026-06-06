package baitan

type PlatformOption struct {
	Value   string  `json:"value"`
	Label   string  `json:"label"`
	Price   float64 `json:"price"`
	DictKey string  `json:"dict_key,omitempty"`
}

func platformOptions() []PlatformOption {
	return []PlatformOption{
		{Value: "1", Label: "工学云"},
		{Value: "2", Label: "校友邦"},
		{Value: "3", Label: "黔职通"},
		{Value: "4", Label: "职校家园"},
		{Value: "5", Label: "习迅云", DictKey: "xxy_school"},
		{Value: "6", Label: "江西职教"},
		{Value: "7", Label: "学习通"},
		{Value: "8", Label: "慧职教", DictKey: "hzj_school"},
		{Value: "9", Label: "广西职业"},
		{Value: "10", Label: "云实习助理", DictKey: "ysx_school"},
		{Value: "11", Label: "习行(学生版)", DictKey: "xxing_school"},
		{Value: "12", Label: "湄洲湾实习"},
		{Value: "13", Label: "校企益邦", DictKey: "xqeb_school"},
		{Value: "14", Label: "博行"},
		{Value: "15", Label: "C30在线"},
		{Value: "16", Label: "智慧教服"},
		{Value: "17", Label: "畅想智行", DictKey: "cxzx_school"},
		{Value: "20", Label: "渤海职业"},
		{Value: "21", Label: "甘肃卫生职业"},
		{Value: "22", Label: "泸州职业"},
		{Value: "23", Label: "数智三职"},
		{Value: "24", Label: "泉州职业(智慧泉海)"},
		{Value: "25", Label: "I鳄院"},
		{Value: "26", Label: "四川机电"},
		{Value: "27", Label: "石家庄职业技术学院"},
		{Value: "28", Label: "苏州职业技术大学"},
		{Value: "29", Label: "惠通江职"},
		{Value: "30", Label: "深圳信息"},
		{Value: "31", Label: "掌上职大"},
		{Value: "32", Label: "乌I职业"},
		{Value: "33", Label: "南京铁道"},
		{Value: "34", Label: "正方教育", DictKey: "zheng_fang"},
		{Value: "35", Label: "安徽林业"},
		{Value: "36", Label: "云上河开"},
		{Value: "37", Label: "喜鹊", DictKey: "xique_school"},
		{Value: "38", Label: "成学云"},
	}
}

func platformLabel(value string) string {
	for _, item := range platformOptions() {
		if item.Value == value {
			return item.Label
		}
	}
	return value
}

func platformDictKey(value string) string {
	if value == "18" {
		return "ykhl_school"
	}
	for _, item := range platformOptions() {
		if item.Value == value {
			return item.DictKey
		}
	}
	return ""
}
