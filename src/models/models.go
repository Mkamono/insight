package models

import (
	"reflect"
	"time"

	"gorm.io/gorm"
)

// Fragment は短い情報の断片を表すモデル
type Fragment struct {
	gorm.Model

	// 基本情報
	Content string `gorm:"type:text;not null"`

	// ドキュメントとの関連
	Documents []Document `gorm:"many2many:document_fragments;"`

	// 関連タグ
	Tags []Tag `gorm:"many2many:fragment_tags;"`
}

// Document は複数のFragmentをまとめて要約したドキュメント
type Document struct {
	gorm.Model

	// 基本情報
	Title   string `gorm:"size:500;not null"`
	Summary string `gorm:"type:text;not null"` // 要約
	Content string `gorm:"type:text;not null"` // Markdown形式のドキュメント

	// バージョン管理
	VersionCreatedAt time.Time `gorm:"not null;index"` // 同一バッチで作成されたドキュメント群のバージョンタイムスタンプ

	// 関連フラグメント
	Fragments []Fragment `gorm:"many2many:document_fragments;"`

	// 関連タグ
	Tags []Tag `gorm:"many2many:document_tags;"`
}

// Tag はドキュメントやフラグメントを分類するためのタグ
type Tag struct {
	gorm.Model

	// 基本情報
	Name  string `gorm:"size:100;unique;not null"`
	Color string `gorm:"size:7"` // hex color code (#RRGGBB)

	// 関連エンティティ
	Documents []Document `gorm:"many2many:document_tags;"`
	Fragments []Fragment `gorm:"many2many:fragment_tags;"`
}

// GetAllModels はこのパッケージ内のすべてのGORMモデルを返します
func GetAllModels() []interface{} {
	return []interface{}{
		&Fragment{},
		&Document{},
		&Tag{},
	}
}

// GetModelNames はモデル名のリストを返します（ログ用）
func GetModelNames() []string {
	models := GetAllModels()
	names := make([]string, len(models))
	for i, model := range models {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		names[i] = t.Name()
	}
	return names
}
