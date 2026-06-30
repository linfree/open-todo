package database

import (
	"testing"
)

func TestCategoryCRUD(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Create
	cat := &Category{
		UserID:   "test-user",
		Name:     "Work",
		Color:    "#ff0000",
		OrderNum: 0,
	}
	if err := db.SaveCategory(cat); err != nil {
		t.Fatalf("save category: %v", err)
	}
	if cat.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Read
	cats, err := db.GetCategories("test-user")
	if err != nil {
		t.Fatalf("get categories: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
	if cats[0].Name != "Work" {
		t.Fatalf("expected name 'Work', got '%s'", cats[0].Name)
	}

	// Update
	cat.Name = "Personal"
	if err := db.SaveCategory(cat); err != nil {
		t.Fatalf("update category: %v", err)
	}
	cats, _ = db.GetCategories("test-user")
	if cats[0].Name != "Personal" {
		t.Fatalf("expected updated name 'Personal', got '%s'", cats[0].Name)
	}

	// Delete
	if err := db.DeleteCategory(cat.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	cats, _ = db.GetCategories("test-user")
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories after delete, got %d", len(cats))
	}
}

func TestTagCRUD(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Create
	tag := &Tag{
		UserID: "test-user",
		Name:   "urgent",
		Color:  "#ff0000",
	}
	if err := db.SaveTag(tag); err != nil {
		t.Fatalf("save tag: %v", err)
	}
	if tag.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Read
	tags, err := db.GetTags("test-user")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "urgent" {
		t.Fatalf("expected name 'urgent', got '%s'", tags[0].Name)
	}

	// Update
	tag.Name = "important"
	if err := db.SaveTag(tag); err != nil {
		t.Fatalf("update tag: %v", err)
	}
	tags, _ = db.GetTags("test-user")
	if tags[0].Name != "important" {
		t.Fatalf("expected updated name 'important', got '%s'", tags[0].Name)
	}

	// Delete
	if err := db.DeleteTag(tag.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	tags, _ = db.GetTags("test-user")
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(tags))
	}
}

func TestGetCategoriesEmpty(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	cats, err := db.GetCategories("nonexistent")
	if err != nil {
		t.Fatalf("get categories: %v", err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(cats))
	}
}

func TestGetTagsEmpty(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	tags, err := db.GetTags("nonexistent")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags, got %d", len(tags))
	}
}
