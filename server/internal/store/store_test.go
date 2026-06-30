package store

import (
	"testing"

	"github.com/linfree/open-todo/server/internal/database"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := database.New("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db)
}

func TestStoreCategoryCRUD(t *testing.T) {
	s := newTestStore(t)

	cat := &Category{
		UserID:   "test-user",
		Name:     "Work",
		Color:    "#ff0000",
		OrderNum: 0,
	}
	if err := s.DB.SaveCategory(cat); err != nil {
		t.Fatalf("save category: %v", err)
	}
	if cat.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	cats, err := s.DB.GetCategories("test-user")
	if err != nil {
		t.Fatalf("get categories: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
	if cats[0].Name != "Work" {
		t.Fatalf("expected name 'Work', got '%s'", cats[0].Name)
	}

	cat.Name = "Personal"
	if err := s.DB.SaveCategory(cat); err != nil {
		t.Fatalf("update category: %v", err)
	}
	cats, _ = s.DB.GetCategories("test-user")
	if cats[0].Name != "Personal" {
		t.Fatalf("expected updated name 'Personal', got '%s'", cats[0].Name)
	}

	if err := s.DB.DeleteCategory(cat.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	cats, _ = s.DB.GetCategories("test-user")
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories after delete, got %d", len(cats))
	}
}

func TestStoreTagCRUD(t *testing.T) {
	s := newTestStore(t)

	tag := &Tag{
		UserID: "test-user",
		Name:   "urgent",
		Color:  "#ff0000",
	}
	if err := s.DB.SaveTag(tag); err != nil {
		t.Fatalf("save tag: %v", err)
	}
	if tag.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	tags, err := s.DB.GetTags("test-user")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "urgent" {
		t.Fatalf("expected name 'urgent', got '%s'", tags[0].Name)
	}

	tag.Name = "important"
	if err := s.DB.SaveTag(tag); err != nil {
		t.Fatalf("update tag: %v", err)
	}
	tags, _ = s.DB.GetTags("test-user")
	if tags[0].Name != "important" {
		t.Fatalf("expected updated name 'important', got '%s'", tags[0].Name)
	}

	if err := s.DB.DeleteTag(tag.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	tags, _ = s.DB.GetTags("test-user")
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(tags))
	}
}

func TestStoreGetCategoriesEmpty(t *testing.T) {
	s := newTestStore(t)
	cats, err := s.DB.GetCategories("nonexistent")
	if err != nil {
		t.Fatalf("get categories: %v", err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(cats))
	}
}

func TestStoreGetTagsEmpty(t *testing.T) {
	s := newTestStore(t)
	tags, err := s.DB.GetTags("nonexistent")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags, got %d", len(tags))
	}
}
