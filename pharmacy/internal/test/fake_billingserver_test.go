package test

import (
	"context"

	billing "github.com/Eucastan/hms/shared/pkg/proto/billing"
)

type FakeBillingServer struct {
	billing.UnimplementedBillingServiceServer
}

func (f *FakeBillingServer) CreateBillCharge(
	ctx context.Context,
	req *billing.CreateBillChargeRequest,
) (*billing.CreateBillChargeResponse, error) {

	// simulate success
	return &billing.CreateBillChargeResponse{
		Status:  "success",
		Message: "bill created",
	}, nil
}
