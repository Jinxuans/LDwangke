package supplier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go-api/internal/model"
)

func (s *Service) QueryCourse(cid int, userinfo string) (*model.CourseQueryResponse, error) {
	cls, err := s.GetClassFull(cid)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	parts := strings.Fields(userinfo)
	var school, user, pass string
	if len(parts) >= 3 {
		school = parts[0]
		user = parts[1]
		pass = parts[2]
	} else if len(parts) == 2 {
		school = "自动识别"
		user = parts[0]
		pass = parts[1]
	} else {
		return nil, errors.New("下单信息格式错误，请输入：学校 账号 密码")
	}

	dockingID := 0
	fmt.Sscanf(cls.Docking, "%d", &dockingID)
	if dockingID == 0 {
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "此课程无需查课，直接下单即可",
			Data:     []model.CourseItem{},
		}, nil
	}

	sup, err := s.GetSupplierByHID(dockingID)
	if err != nil {
		return nil, err
	}

	cfg := GetPlatformConfig(sup.PT)
	switch cfg.QueryAct {
	case "local_time":
		data := s.generateLocalTimeList(cls.Noun)
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     data,
		}, nil
	case "local_script":
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      fmt.Sprintf("平台 %s 暂不支持自动查课，请直接下单", sup.PT),
			Data:     []model.CourseItem{},
		}, nil
	case "xxt_query":
		result, err := xxtCallQuery(user, pass, school)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("学习通查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		if codeVal, _ := result["code"].(int); codeVal == -1 {
			msg, _ := result["msg"].(string)
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      msg,
				Data:     []model.CourseItem{},
			}, nil
		}

		var items []model.CourseItem
		switch data := result["data"].(type) {
		case []map[string]interface{}:
			for _, d := range data {
				items = append(items, model.CourseItem{
					ID:   toString(d["id"]),
					Name: toString(d["name"]),
				})
			}
		case []interface{}:
			for _, item := range data {
				if d, ok := item.(map[string]interface{}); ok {
					items = append(items, model.CourseItem{
						ID:   toString(d["id"]),
						Name: toString(d["name"]),
					})
				}
			}
		}

		if userInfoStr, ok := result["userinfo"].(string); ok {
			userinfo = userInfoStr
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     items,
		}, nil
	case "KUN_custom":
		result, err := kunCallQuery(sup, cls.Noun, school, user, pass)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      result.Msg,
			Data:     result.Data,
		}, nil
	case "simple_custom":
		result, err := simpleCallQuery(sup, cls.Noun, school, user, pass)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: result.UserName,
			Msg:      result.Msg,
			Data:     result.Data,
		}, nil
	case "yyy_custom":
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     []model.CourseItem{},
		}, nil
	default:
		result, err := s.callSupplierQuery(sup, cls, school, user, pass)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: result.UserName,
			Msg:      result.Msg,
			Data:     result.Data,
		}, nil
	}
}

func (s *Service) generateLocalTimeList(noun string) []model.CourseItem {
	hoursPerUnit := 6
	if noun == "1" {
		hoursPerUnit = 5
	}
	items := make([]model.CourseItem, 0, 20)
	for i := 1; i <= 20; i++ {
		total := i * hoursPerUnit
		items = append(items, model.CourseItem{
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("从第一个开始选择，每选中一个加%d小时，选到此处总时长为 %d 小时", hoursPerUnit, total),
		})
	}
	return items
}

func (s *Service) callSupplierQuery(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass string) (*model.SupplierQueryResult, error) {
	apiURL := buildSupplierURL(sup.URL, "get")

	values := url.Values{}
	values.Set("uid", sup.User)
	values.Set("key", sup.Pass)
	values.Set("school", school)
	values.Set("user", user)
	values.Set("pass", pass)
	values.Set("platform", cls.Noun)

	resp, err := s.client.PostForm(apiURL, values)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	msg, _ := raw["msg"].(string)
	userName, _ := raw["userName"].(string)

	var items []model.CourseItem
	if dataArr, ok := raw["data"].([]interface{}); ok {
		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, model.CourseItem{
					ID:             toString(m["id"]),
					Name:           toString(m["name"]),
					KCJS:           toString(m["kcjs"]),
					StudyStartTime: toString(m["studyStartTime"]),
					StudyEndTime:   toString(m["studyEndTime"]),
					ExamStartTime:  toString(m["examStartTime"]),
					ExamEndTime:    toString(m["examEndTime"]),
					Complete:       toString(m["complete"]),
				})
			}
		}
	}

	return &model.SupplierQueryResult{
		Msg:      msg,
		UserName: userName,
		Data:     items,
	}, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
