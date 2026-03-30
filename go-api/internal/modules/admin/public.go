package admin

import (
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var publicConfigKeys = map[string]bool{
	"sitename": true, "logo": true, "hlogo": true,
	"sykg": true, "bz": true, "notice": true, "tcgonggao": true,
	// 独立的首页渠道公告内容，避免和登录弹窗公告共用同一个键。
	"qd_notice": true,
	// 消费排行榜开关：默认关闭，只有显式为 1 时才开启。
	"top_consumers_open": true,
	"flkg":               true, "fllx": true, "fontsZDY": true, "fontsFamily": true,
	"qd_notice_open": true, "xdsmopen": true, "anti_debug": true,
	"version": true, "onlineStore_trdltz": true, "sjqykg": true,
	"user_yqzc": true, "login_slider_verify": true, "login_email_verify": true,
	"webVfx_open": true, "webVfx": true,
	"keywords": true, "description": true,
	"checkin_enabled":      true,
	"recharge_bonus_rules": true,
	"pass2_kg":             true,
	"recommend_channels":   true,
	"login_forget_pwd":     true,
}

func getTopConsumers(period string) []map[string]interface{} {
	var interval string
	switch period {
	case "week":
		interval = "INTERVAL 6 DAY"
	case "month":
		interval = "INTERVAL 29 DAY"
	default:
		interval = "INTERVAL 0 DAY"
	}

	query := "SELECT o.uid, COALESCE(u.user,''), COALESCE(u.faceimg,''), COUNT(*), COALESCE(SUM(o.fees),0) " +
		"FROM qingka_wangke_order o LEFT JOIN qingka_wangke_user u ON o.uid=u.uid " +
		"WHERE o.addtime >= CURDATE() - " + interval + " " +
		"GROUP BY o.uid ORDER BY SUM(o.fees) DESC LIMIT 10"

	rows, err := database.DB.Query(query)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var uid, cnt int
		var username, faceimg string
		var total float64
		rows.Scan(&uid, &username, &faceimg, &cnt, &total)

		avatar := faceimg
		if avatar == "" && username != "" {
			avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + username + "&s=40"
		}

		list = append(list, map[string]interface{}{
			"uid":      uid,
			"username": username,
			"avatar":   avatar,
			"orders":   cnt,
			"total":    total,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list
}

func getPublicConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config")
	if err != nil {
		return map[string]string{}, nil
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		if publicConfigKeys[k] {
			result[k] = v
		}
	}
	return result, nil
}

func listPublicAnnouncements(uid int, page, limit int) ([]model.Announcement, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	var total int64
	var parentID int
	if uid > 0 {
		database.DB.QueryRow("SELECT COALESCE(uuid, 0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&parentID)
	}

	where := "status = '1' AND (visibility = 0 OR (visibility = 1 AND uid = ?))"
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_gonggao WHERE "+where, parentID).Scan(&total)

	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT id, COALESCE(title,''), COALESCE(content,''), COALESCE(time,''), COALESCE(uid,0), COALESCE(status,'1'), COALESCE(zhiding,'0'), COALESCE(author,''), COALESCE(visibility,0) FROM qingka_wangke_gonggao WHERE "+where+" ORDER BY zhiding DESC, id DESC LIMIT ? OFFSET ?",
		parentID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Announcement
	for rows.Next() {
		var a model.Announcement
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.Time, &a.UID, &a.Status, &a.Zhiding, &a.Author, &a.Visibility)
		list = append(list, a)
	}
	if list == nil {
		list = []model.Announcement{}
	}
	return list, total, nil
}

func TopConsumers(c *gin.Context) {
	// 消费排行榜默认关闭：只有显式配置为 1 时才返回排行数据。
	var rankingEnabled string
	_ = database.DB.QueryRow("SELECT COALESCE(`k`,'') FROM qingka_wangke_config WHERE `v` = ?", "top_consumers_open").Scan(&rankingEnabled)
	if rankingEnabled != "1" {
		response.Success(c, []map[string]interface{}{})
		return
	}

	period := c.DefaultQuery("period", "day")
	response.Success(c, getTopConsumers(period))
}

func SiteConfigGet(c *gin.Context) {
	config, err := getPublicConfig()
	if err != nil {
		response.ServerErrorf(c, err, "查询站点配置失败")
		return
	}
	response.Success(c, config)
}

func AnnouncementListPublic(c *gin.Context) {
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, total, err := listPublicAnnouncements(uid, page, limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询公告失败")
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}
