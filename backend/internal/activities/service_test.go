package activities

import (
	"context"
	"testing"
)

func TestServiceRecordsAndListsActivityCounts(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.RecordPostView(ctx, "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439021"); err != nil {
		t.Fatalf("record view failed: %v", err)
	}
	if err := svc.RecordPostView(ctx, "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439021"); err != nil {
		t.Fatalf("record second view failed: %v", err)
	}
	if err := svc.RecordPostLike(ctx, "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439021"); err != nil {
		t.Fatalf("record like failed: %v", err)
	}
	if err := svc.RecordComment(ctx, "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439021", "507f1f77bcf86cd799439031"); err != nil {
		t.Fatalf("record comment failed: %v", err)
	}

	counts, err := svc.Counts(ctx, "507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatalf("counts failed: %v", err)
	}
	if counts.Views != 1 || counts.Likes != 1 || counts.Comments != 1 {
		t.Fatalf("expected unique counts, got %+v", counts)
	}

	views, err := svc.List(ctx, "507f1f77bcf86cd799439011", ActivityTypeView, 20)
	if err != nil {
		t.Fatalf("list views failed: %v", err)
	}
	if len(views) != 1 || views[0].Count != 2 {
		t.Fatalf("expected one view activity with count 2, got %+v", views)
	}
}
