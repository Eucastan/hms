package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	billing "github.com/Eucastan/hms/shared/pkg/proto/billing"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
	"github.com/Eucastan/hms/shared/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	billingClient billing.BillingServiceClient
	patientClient patient.PatientServiceClient
}

// For mock connection
var _ ServiceClient = (*Clients)(nil)

func dial(Addr string) *grpc.ClientConn {
	retryCfg, err := os.ReadFile("retry_policy.json")

	Opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithDefaultServiceConfig(string(retryCfg)),
	}

	if err == nil {
		Opts = append(Opts, grpc.WithDefaultServiceConfig(string(retryCfg)))
	}

	Conn, err := grpc.NewClient(Addr, Opts...)
	if err != nil {
		log.Printf("Failed to connect to Payment server: %v", err)
	}

	return Conn
}

// Production constructor
func NewClients(billingAddr, patientAddr string) *Clients {
	return &Clients{
		billingClient: billing.NewBillingServiceClient(dial(billingAddr)),
		patientClient: patient.NewPatientServiceClient(dial(patientAddr)),
	}
}

// Test constructor - For testing purposes
func NewTestClients(billingClient billing.BillingServiceClient, patientClient patient.PatientServiceClient) *Clients {
	return &Clients{
		billingClient: billingClient,
		patientClient: patientClient,
	}
}

func (c *Clients) CreateBillCharge(
	ctx context.Context,
	patientID,
	sourceRefId,
	description string,
	qty int32,
	total float64,
) error {

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	_, err := c.billingClient.CreateBillCharge(
		ctx,
		&billing.CreateBillChargeRequest{
			PatientId:     patientID,
			SourceRefId:   sourceRefId,
			ReferenceType: "Drug",
			Description:   description,
			Quantity:      qty,
			TotalPrice:    total,
		},
	)

	return err
}

func (c *Clients) ValidatePatientID(ctx context.Context, patientID string) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req := &patient.GetPatientRequest{
		PatientId: patientID,
	}

	resp, err := c.patientClient.GetPatient(ctx, req)
	if err != nil {
		return utils.ToPublicError(err)
	}

	if resp.PatientInfo == nil {
		return fmt.Errorf("Patient not found")
	}

	return nil
}
