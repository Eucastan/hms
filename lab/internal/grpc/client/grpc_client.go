package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
	"github.com/Eucastan/hms/shared/pkg/utils"
)

type Clients struct {
	patientClient patient.PatientServiceClient
	retryConfig   string
}

func dial(Addr, retryConfig string) *grpc.ClientConn {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if retryConfig != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(retryConfig))
	}

	conn, err := grpc.NewClient(Addr)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", Addr, err)
	}

	return conn
}

func SendClient(patientAddr string) *Clients {
	retryCfgBytes, err := os.ReadFile("retry_policy.json")
	retryConfig := ""

	if err != nil {
		log.Printf("Warning: could not load retry-policy: %v", err)
	} else {
		retryConfig = string(retryCfgBytes)
	}

	return &Clients{
		patientClient: patient.NewPatientServiceClient(dial(patientAddr, retryConfig)),
	}
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
