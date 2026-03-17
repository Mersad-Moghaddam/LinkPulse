package repository

import (
	"context"
	"testing"

	"linkpulse/internal/models"
)

func TestDeleteByCodeClearsAnalyticsState(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	link := models.Link{ID: "1", ShortCode: "reuse01", LongURL: "https://a.example"}
	if _, err := repo.Create(ctx, link); err != nil {
		t.Fatal(err)
	}
	_ = repo.Record(ctx, models.Click{LinkID: "reuse01", IP: "1.1.1.1", UserAgent: "ua", Referrer: "r1", Browser: "chrome"})
	_ = repo.Record(ctx, models.Click{LinkID: "reuse01", IP: "1.1.1.1", UserAgent: "ua", Referrer: "r1", Browser: "chrome"})

	if err := repo.DeleteByCode(ctx, "reuse01"); err != nil {
		t.Fatal(err)
	}

	// recreate same alias; analytics must not inherit old click/visited state
	link2 := models.Link{ID: "2", ShortCode: "reuse01", LongURL: "https://b.example"}
	if _, err := repo.Create(ctx, link2); err != nil {
		t.Fatal(err)
	}
	_ = repo.Record(ctx, models.Click{LinkID: "reuse01", IP: "1.1.1.1", UserAgent: "ua", Referrer: "r2", Browser: "firefox"})

	s, err := repo.SummaryByCode(ctx, "reuse01")
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalClicks != 1 {
		t.Fatalf("expected 1 click after alias reuse, got %d", s.TotalClicks)
	}
	if s.UniqueClicks != 1 {
		t.Fatalf("expected unique count reset after delete/recreate, got %d", s.UniqueClicks)
	}
}
