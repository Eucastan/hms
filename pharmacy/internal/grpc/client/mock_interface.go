package client

import "context"

type ServiceClient interface {
	CreateBillCharge(
		ctx context.Context,
		patientID,
		sourceRefId,
		description string,
		qty int32,
		total float64,
	) error

	ValidatePatientID(ctx context.Context, patientID string) error
}