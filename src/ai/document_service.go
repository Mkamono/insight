package ai

import (
	"context"

	"gorm.io/gorm"
)

// Service はAI機能を提供するファサードサービス
type Service struct {
	db *gorm.DB
}

// NewService は新しいServiceを作成
func NewService(db *gorm.DB) (*Service, error) {
	return &Service{
		db: db,
	}, nil
}

// CreateDocuments はフラグメントからドキュメントを作成
func (s *Service) CreateDocuments(ctx context.Context) error {
	generator, err := NewDocumentGenerator(s.db)
	if err != nil {
		return err
	}
	return generator.GenerateDocuments(ctx)
}

// CompressFragments はフラグメントを圧縮する
func (s *Service) CompressFragments(ctx context.Context) error {
	compressor, err := NewFragmentCompressor(s.db)
	if err != nil {
		return err
	}
	return compressor.CompressFragments(ctx)
}

