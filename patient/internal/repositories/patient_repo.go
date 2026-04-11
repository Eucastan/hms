package repositories

import (
	"context"
	"errors"

	"github.com/Eucastan/hms/patient/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *models.Patient) (*models.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error)
	FindByNameAndHospitalNo(ctx context.Context, name, hospitalNo string) ([]*models.Patient, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PatientRepo struct {
	DB *gorm.DB
}

var ErrNotFound = errors.New("record not found")

func NewPatientRepo(db *gorm.DB) PatientRepository {
	return &PatientRepo{DB: db}
}

func (r *PatientRepo) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	err := r.DB.WithContext(ctx).Create(patient).Error
	if err != nil {
		return nil, err
	}
	return patient, nil
}

func (r *PatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error) {
	var pt models.Patient
	err := r.DB.WithContext(ctx).
		Preload("Admission").
		Where("id = ?", id).
		First(&pt).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &pt, err
}

func (r *PatientRepo) FindByNameAndHospitalNo(ctx context.Context, name, hospitalNo string) ([]*models.Patient, error) {
	var pts []*models.Patient
	query := r.DB.WithContext(ctx)

	if hospitalNo != "" {
		query = query.Where("hospital_no ILIKE ?", "%"+hospitalNo+"%")
	}
	if name != "" {
		query = query.Where("(firstname ILIKE ? OR lastname ILIKE ?)", "%"+name+"%", "%"+name+"%")
	}

	err := query.
		Order("firstname ASC, lastname ASC").
		Limit(20).
		Find(&pts).Error

	return pts, err
}

func (r *PatientRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	// Prevent changing critical fields
	for _, key := range []string{"id", "hospital_no", "created_at", "updated_at", "deleted_at"} {
		delete(updates, key)
	}

	return r.DB.WithContext(ctx).
		Model(&models.Patient{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *PatientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Patient{}).Error
}
