package service

import (
	"fmt"

	"go-api/internal/model"
	ordermodule "go-api/internal/modules/order"
)

type TenantService struct{}

var tenantMallOrderAdder = func(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	return ordermodule.NewServices().Command.AddForMall(bUID, tid, cUID, retailPrice, req)
}

func buildMallOrderAddRequest(order model.MallPayOrder) model.OrderAddRequest {
	userInfo := fmt.Sprintf("自动识别 %s %s", order.Account, order.Password)
	item := model.OrderAddItem{UserInfo: userInfo}
	if order.CourseID != "" {
		item.Data = model.OrderAddCourse{
			ID:   order.CourseID,
			Name: order.CourseName,
			KCJS: order.CourseKCJS,
		}
	}

	return model.OrderAddRequest{
		CID:  order.CID,
		Data: []model.OrderAddItem{item},
	}
}

func submitMallOrder(bUID, tid int, order model.MallPayOrder) (*model.OrderAddResult, error) {
	return tenantMallOrderAdder(bUID, tid, order.CUID, order.Money, buildMallOrderAddRequest(order))
}
