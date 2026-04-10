package test

import (
	"context"
)

type MockClient struct{}

func (m *MockClient) CreateBillCharge(
	ctx context.Context,
	patientID,
	sourceRefId,
	description string,
	qty int32,
	total float64,
) error {
	return nil // simulate success
}

func (m *MockClient) ValidatePatientID(ctx context.Context, patientID string) error {
	return nil // always valid
}
