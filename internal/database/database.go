package database

import (
	"fmt"
	"log"
	"time"

	"survey2earn-backend/internal/config"
	"survey2earn-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()
	
	var gormLogger logger.Interface
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	log.Println("Database connected successfully")
	
	return &Database{DB: db}, nil
}

func (d *Database) AutoMigrate() error {
	log.Println("Running database migrations...")
	
	err := d.DB.AutoMigrate(
		&models.User{},
		&models.AuthSession{},
		&models.UserStats{},
		&models.UserBalance{},
		
		&models.Survey{},
		&models.Question{},
		
		&models.Response{},
		&models.Answer{},
		&models.ResponseSummary{},
		
		&models.RewardPool{},
		&models.RewardTransaction{},
		&models.WithdrawalRequest{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	if err := d.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	if err := d.seedData(); err != nil {
		log.Printf("Warning: failed to seed data: %v", err)
	}
	
	log.Println("Database migrations completed successfully")
	return nil
}

func (d *Database) createIndexes() error {
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_wallet_lower ON users(LOWER(wallet_address))")
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_reputation ON users(reputation_score DESC)")
	
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_surveys_status_category ON surveys(status, category)")
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_surveys_created_at ON surveys(created_at DESC)")
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_surveys_active ON surveys(status, start_date, end_date)")
	
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_responses_user_survey ON responses(user_id, survey_id)")
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_responses_completed_at ON responses(completed_at DESC)")
	
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_status_created ON reward_transactions(status, created_at)")
	d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_user_status ON reward_transactions(user_id, status)")
	
	return nil
}

func (d *Database) seedData() error {
	var count int64
	d.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil 
	}
	
	categories := []string{
		"Technology",
		"Finance",
		"Healthcare", 
		"Education",
		"Entertainment",
		"Gaming",
		"DeFi",
		"NFT",
		"AI/ML",
		"General",
	}
	
	log.Printf("Available survey categories: %v", categories)
	
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (d *Database) GetStats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
	}
}