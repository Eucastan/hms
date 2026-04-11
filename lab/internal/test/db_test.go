package test

import (
	"context"
	"testing"
	"time"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("lab_test_db"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort("5432/tcp"),
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			).WithDeadline(60*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "failed to terminate container")
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	db.Debug()

	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error
	require.NoError(t, err, "failed to enable uuid extension")

	err = db.AutoMigrate(&models.LabTestRequest{}, &models.LabResult{})
	require.NoError(t, err, "failed to migrate schema")

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE lab_test_requests, lab_results RESTART IDENTITY CASCADE")
	})

	return db
}
