// Package plugins 通过 blank import 触发各插件的 init() 自注册。
// 新增插件时，只需在此文件添加一行 import 即可。
package plugins

import (
	_ "go-api/internal/plugins/appui"
	_ "go-api/internal/plugins/paper"
	_ "go-api/internal/plugins/sdxy"
	_ "go-api/internal/plugins/sxdk"
	_ "go-api/internal/plugins/tuboshu"
	_ "go-api/internal/plugins/tutuqg"
	_ "go-api/internal/plugins/tuzhi"
	_ "go-api/internal/plugins/w"
	_ "go-api/internal/plugins/xm"
	_ "go-api/internal/plugins/ydsj"
	_ "go-api/internal/plugins/yfdk"
	_ "go-api/internal/plugins/yongye"
)
