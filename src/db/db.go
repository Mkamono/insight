package db

import (
	"insight/src/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config はデータベース設定
type Config struct {
	DatabasePath string
	LogLevel     logger.LogLevel
}

// DefaultConfig はデフォルトのデータベース設定
func DefaultConfig() *Config {
	return &Config{
		DatabasePath: "insight.db",
		LogLevel:     logger.Silent,
	}
}

// Init はデータベースを初期化する
func Init(config *Config) (*gorm.DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// GORM設定
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}

	// SQLiteデータベースに接続
	db, err := gorm.Open(sqlite.Open(config.DatabasePath), gormConfig)
	if err != nil {
		return nil, err
	}

	// 自動マイグレーション実行
	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate はデータベースマイグレーションを実行する
func Migrate(db *gorm.DB) error {
	// すべてのモデルを自動取得してマイグレーション実行
	allModels := models.GetAllModels()

	err := db.AutoMigrate(allModels...)
	if err != nil {
		return err
	}

	return nil
}

// Close はデータベース接続を閉じる
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
