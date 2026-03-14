package service

func getAdminConfig() map[string]string {
	conf, _ := GetAdminConfigMap()
	return conf
}

func getAdminConfigValue(key string) string {
	return getAdminConfig()[key]
}

func adminConfigEnabled(key string) bool {
	return getAdminConfigValue(key) == "1"
}

func getConfiguredSiteName() string {
	siteName := getAdminConfigValue("sitename")
	if siteName == "" {
		siteName = "System"
	}
	return siteName
}
