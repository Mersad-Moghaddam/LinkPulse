package repository

import (
	"context"

	"linkpulse/internal/models"
)

var (
	ErrNotFound     = errString("not found")
	ErrAliasExists  = errString("alias already exists")
	ErrUnauthorized = errString("unauthorized")
)

type errString string

func (e errString) Error() string { return string(e) }

type LinkRepository interface {
	Create(ctx context.Context, link models.Link) (models.Link, error)
	GetByCode(ctx context.Context, code string) (models.Link, error)
	List(ctx context.Context) ([]models.Link, error)
	DeleteByCode(ctx context.Context, code string) error
}

type ClickRepository interface {
	Record(ctx context.Context, click models.Click) error
	SummaryByCode(ctx context.Context, code string) (models.AnalyticsSummary, error)
}
