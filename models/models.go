package models

import (
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"

	"github.com/gofrs/uuid"
)

// Model is a customized short-hand for gorm.Model
type Model struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// ModelManager is how we manipulate the various models in lens
type ModelManager struct {
	DBM *database.DatabaseManager
}

// NewModelManager is used to generate our model manager
func NewModelManager(cfg *config.TemporalConfig, runCustomMigrations bool) (*ModelManager, error) {
	dbm, err := database.Initialize(cfg, database.DatabaseOptions{})
	if err != nil {
		return nil, err
	}
	mm := &ModelManager{DBM: dbm}
	if runCustomMigrations {
		if err := mm.RunMigrations(); err != nil {
			return nil, err
		}
	}
	return mm, nil
}

// RunMigrations is used to migrate our custom models
func (mm *ModelManager) RunMigrations() error {
	if err := mm.DBM.DB.AutoMigrate(&Object{}).Error; err != nil {
		return err
	}
	return mm.DBM.DB.AutoMigrate(&MetaData{}).Error
}
