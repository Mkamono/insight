package usecase

import (
	"insight/src/models"

	"gorm.io/gorm"
)

type FragmentUsecase struct {
	db *gorm.DB
}

func NewFragmentUsecase(db *gorm.DB) *FragmentUsecase {
	return &FragmentUsecase{db: db}
}

// CreateFragmentInput はFragment作成の入力データ
type CreateFragmentInput struct {
	Content string `json:"content" validate:"required"`
}

// CreateFragment は新しいFragmentを作成する
func (u *FragmentUsecase) CreateFragment(input CreateFragmentInput) (*models.Fragment, error) {
	fragment := models.Fragment{
		Content: input.Content,
	}

	if err := u.db.Create(&fragment).Error; err != nil {
		return nil, err
	}

	return &fragment, nil
}

// GetFragment はIDでFragmentを取得する
func (u *FragmentUsecase) GetFragment(id uint) (*models.Fragment, error) {
	var fragment models.Fragment
	if err := u.db.First(&fragment, id).Error; err != nil {
		return nil, err
	}
	return &fragment, nil
}

// GetAllFragments はすべてのFragmentを取得する
func (u *FragmentUsecase) GetAllFragments() ([]models.Fragment, error) {
	var fragments []models.Fragment
	if err := u.db.Preload("Tags").Order("created_at DESC").Find(&fragments).Error; err != nil {
		return nil, err
	}
	return fragments, nil
}

// DeleteFragment はFragmentを削除する（ソフトデリート）
func (u *FragmentUsecase) DeleteFragment(id uint) error {
	return u.db.Delete(&models.Fragment{}, id).Error
}
