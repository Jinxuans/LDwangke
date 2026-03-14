package service

import (
	"testing"

	"go-api/internal/model"
)

func TestBuildMallOrderAddRequest(t *testing.T) {
	order := model.MallPayOrder{
		CID:        123,
		Account:    "alice",
		Password:   "secret",
		CourseID:   "course-1",
		CourseName: "Course Name",
		CourseKCJS: "Teacher",
	}

	req := buildMallOrderAddRequest(order)
	if req.CID != order.CID {
		t.Fatalf("unexpected cid: got %d want %d", req.CID, order.CID)
	}
	if len(req.Data) != 1 {
		t.Fatalf("unexpected data length: got %d want 1", len(req.Data))
	}
	if req.Data[0].UserInfo != "自动识别 alice secret" {
		t.Fatalf("unexpected user info: %q", req.Data[0].UserInfo)
	}
	if req.Data[0].Data.ID != order.CourseID || req.Data[0].Data.Name != order.CourseName || req.Data[0].Data.KCJS != order.CourseKCJS {
		t.Fatalf("unexpected course data: %+v", req.Data[0].Data)
	}
}

func TestSubmitMallOrderDelegatesToOrderModule(t *testing.T) {
	oldAdder := tenantMallOrderAdder
	defer func() {
		tenantMallOrderAdder = oldAdder
	}()

	var (
		gotBUID        int
		gotTID         int
		gotCUID        int
		gotRetailPrice float64
		gotReq         model.OrderAddRequest
	)
	tenantMallOrderAdder = func(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
		gotBUID = bUID
		gotTID = tid
		gotCUID = cUID
		gotRetailPrice = retailPrice
		gotReq = req
		return &model.OrderAddResult{SuccessCount: 1, OIDs: []int64{99}}, nil
	}

	order := model.MallPayOrder{
		CID:        7,
		CUID:       8,
		Money:      19.99,
		Account:    "bob",
		Password:   "pwd",
		CourseID:   "cid-1",
		CourseName: "Test Course",
		CourseKCJS: "Desc",
	}

	result, err := submitMallOrder(101, 202, order)
	if err != nil {
		t.Fatalf("submitMallOrder returned error: %v", err)
	}
	if result == nil || result.SuccessCount != 1 || len(result.OIDs) != 1 || result.OIDs[0] != 99 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if gotBUID != 101 || gotTID != 202 || gotCUID != order.CUID {
		t.Fatalf("unexpected delegation args: bUID=%d tid=%d cUID=%d", gotBUID, gotTID, gotCUID)
	}
	if gotRetailPrice != order.Money {
		t.Fatalf("unexpected retail price: got %.2f want %.2f", gotRetailPrice, order.Money)
	}
	if gotReq.CID != order.CID || len(gotReq.Data) != 1 {
		t.Fatalf("unexpected request: %+v", gotReq)
	}
	if gotReq.Data[0].UserInfo != "自动识别 bob pwd" {
		t.Fatalf("unexpected user info: %q", gotReq.Data[0].UserInfo)
	}
}
