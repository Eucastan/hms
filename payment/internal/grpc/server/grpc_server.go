package server

import (
	"context"
	"errors"
	"log"

	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/payment/internal/services"
	billing "github.com/Eucastan/hms/shared/pkg/proto/billing"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingServiceServer struct {
	billing.UnimplementedBillingServiceServer
	svc services.BillingService
}

func NewBillingServer(svc services.BillingService) *BillingServiceServer {
	return &BillingServiceServer{
		svc: svc,
	}
}

func (s *BillingServiceServer) CreateBillCharge(ctx context.Context, req *billing.CreateBillChargeRequest) (*billing.CreateBillChargeResponse, error) {
	createdBy := ctx.Value("user_id").(string)

	if req.PatientId == "" || req.SourceRefId == "" || req.ReferenceType == "" || req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}
	if req.TotalPrice <= 0 {
		return nil, status.Error(codes.InvalidArgument, "unit price must be positive")
	}

	billCharge := models.BillChargeRequest{
		PatientID:     req.PatientId,
		SourceRefID:   req.SourceRefId,
		ReferenceType: req.ReferenceType,
		Description:   req.Description,
		Quantity:      req.Quantity,
		UnitPrice:     req.TotalPrice,
		CreatedBy:     createdBy,
	}
	charge, invoice, total, err := s.svc.CreateBillCharge(ctx, billCharge)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &billing.CreateBillChargeResponse{
		BillChargeId: charge.ID.String(),
		InvoiceId:    invoice.ID.String(),
		InvoiceTotal: total,
		Status:       string(charge.Status),
		Message:      "Bill charge created successfully",
	}, nil
}

func (s *BillingServiceServer) RefundBillCharge(ctx context.Context, req *billing.RefundBillChargeRequest) (*billing.RefundBillChargeResponse, error) {
	refundedBy, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid charge_id")
	}

	chargeID, err := uuid.Parse(req.ChargeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid charge_id")
	}

	err = s.svc.RefundBillCharge(ctx, chargeID, refundedBy)
	if err != nil {
		if errors.Is(err, errors.New("you can only refund charges you created")) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		log.Printf("Refund failed for charge %s: %v", req.ChargeId, err)
		return nil, status.Error(codes.Internal, "failed to refund charge")
	}

	return &billing.RefundBillChargeResponse{
		Status:  "refunded",
		Message: "Charge refunded successfully",
	}, nil
}
