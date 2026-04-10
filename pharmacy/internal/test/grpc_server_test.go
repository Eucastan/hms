package test

import (
	"net"
	"testing"

	"google.golang.org/grpc"

	billing "github.com/Eucastan/hms/shared/pkg/proto/billing"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
)

type GRPCTestServer struct {
	Listener net.Listener
	Server   *grpc.Server
	Addr     string
}

func StartGRPCTestServer(t *testing.T) *GRPCTestServer {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // random port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// register fake services
	billing.RegisterBillingServiceServer(s, &FakeBillingServer{})
	patient.RegisterPatientServiceServer(s, &FakePatientServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("grpc server failed: %v", err)
		}
	}()

	return &GRPCTestServer{
		Listener: lis,
		Server:   s,
		Addr:     lis.Addr().String(),
	}
}

func (g *GRPCTestServer) Close() {
	g.Server.Stop()
	g.Listener.Close()
}
