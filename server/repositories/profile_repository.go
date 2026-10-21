package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProfileRepository interface {
	Create(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)
	Update(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
	Upsert(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *profileRepository) Create(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Create(profile).Error
}

func (r *profileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Update(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Save(profile).Error
}

func (r *profileRepository) Upsert(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"date_of_birth", "gender", "phone_number", "address", "preferences", "updated_at"}),
	}).Create(profile).Error
}
