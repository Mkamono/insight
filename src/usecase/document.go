package usecase

import (
	"insight/src/models"
	"time"

	"gorm.io/gorm"
)

type DocumentUsecase struct {
	db *gorm.DB
}

func NewDocumentUsecase(db *gorm.DB) *DocumentUsecase {
	return &DocumentUsecase{db: db}
}

// CreateDocumentInput はDocument作成の入力データ
type CreateDocumentInput struct {
	Title            string    `json:"title" validate:"required"`
	Summary          string    `json:"summary" validate:"required"`
	Content          string    `json:"content" validate:"required"`
	VersionCreatedAt time.Time `json:"version_created_at" validate:"required"`
	FragmentIDs      []uint    `json:"fragment_ids"`
	TagIDs           []uint    `json:"tag_ids"`
}

// CreateDocument は新しいDocumentを作成する
func (u *DocumentUsecase) CreateDocument(input CreateDocumentInput) (*models.Document, error) {
	document := models.Document{
		Title:            input.Title,
		Summary:          input.Summary,
		Content:          input.Content,
		VersionCreatedAt: input.VersionCreatedAt,
	}

	// トランザクション内で実行
	err := u.db.Transaction(func(tx *gorm.DB) error {
		// Documentを作成
		if err := tx.Create(&document).Error; err != nil {
			return err
		}

		// Fragmentとの関連を設定
		if len(input.FragmentIDs) > 0 {
			var fragments []models.Fragment
			if err := tx.Find(&fragments, input.FragmentIDs).Error; err != nil {
				return err
			}

			// 多対多の関連を設定
			if err := tx.Model(&document).Association("Fragments").Append(fragments); err != nil {
				return err
			}
		}

		// Tagとの関連を設定
		if len(input.TagIDs) > 0 {
			var tags []models.Tag
			if err := tx.Find(&tags, input.TagIDs).Error; err != nil {
				return err
			}

			// 多対多の関連を設定
			if err := tx.Model(&document).Association("Tags").Append(tags); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &document, nil
}

// GetDocument はIDでDocumentを取得する（関連FragmentとTagも含む）
func (u *DocumentUsecase) GetDocument(id uint) (*models.Document, error) {
	var document models.Document
	if err := u.db.Preload("Fragments").Preload("Tags").First(&document, id).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

// GetAllDocuments はすべてのDocumentを取得する
func (u *DocumentUsecase) GetAllDocuments() ([]models.Document, error) {
	var documents []models.Document
	if err := u.db.Preload("Fragments").Preload("Tags").Order("version_created_at DESC, created_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetDocumentsByVersion は指定されたバージョンのDocumentを取得する
func (u *DocumentUsecase) GetDocumentsByVersion(versionCreatedAt time.Time) ([]models.Document, error) {
	var documents []models.Document
	if err := u.db.Preload("Fragments").Preload("Tags").Where("version_created_at = ?", versionCreatedAt).Order("created_at ASC").Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetDistinctVersions は全ての異なるバージョンタイムスタンプを取得する
func (u *DocumentUsecase) GetDistinctVersions() ([]time.Time, error) {
	var versions []time.Time
	if err := u.db.Model(&models.Document{}).Select("DISTINCT version_created_at").Order("version_created_at DESC").Pluck("version_created_at", &versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

// AddFragmentToDocument は既存のDocumentにFragmentを追加する
func (u *DocumentUsecase) AddFragmentToDocument(documentID, fragmentID uint) error {
	var document models.Document
	var fragment models.Fragment

	return u.db.Transaction(func(tx *gorm.DB) error {
		// Documentの存在確認
		if err := tx.First(&document, documentID).Error; err != nil {
			return err
		}

		// Fragmentの存在確認
		if err := tx.First(&fragment, fragmentID).Error; err != nil {
			return err
		}

		// 関連を追加
		return tx.Model(&document).Association("Fragments").Append(&fragment)
	})
}

// RemoveFragmentFromDocument は DocumentからFragmentを削除する
func (u *DocumentUsecase) RemoveFragmentFromDocument(documentID, fragmentID uint) error {
	var document models.Document
	var fragment models.Fragment

	return u.db.Transaction(func(tx *gorm.DB) error {
		// Documentの存在確認
		if err := tx.First(&document, documentID).Error; err != nil {
			return err
		}

		// Fragmentの存在確認
		if err := tx.First(&fragment, fragmentID).Error; err != nil {
			return err
		}

		// 関連を削除
		return tx.Model(&document).Association("Fragments").Delete(&fragment)
	})
}

// DeleteDocument はDocumentを削除する（ソフトデリート）
func (u *DocumentUsecase) DeleteDocument(id uint) error {
	return u.db.Delete(&models.Document{}, id).Error
}
