package repository

import (
	"context"

	"github.com/elchemista/driplnk/internal/domain"
)

// PebbleLinkRepository wraps PebbleRepository to implement domain.LinkRepository.
type PebbleLinkRepository struct {
	repo *PebbleRepository
}

func NewPebbleLinkRepository(repo *PebbleRepository) *PebbleLinkRepository {
	return &PebbleLinkRepository{repo: repo}
}

func (r *PebbleLinkRepository) Save(ctx context.Context, link *domain.Link) error {
	return r.repo.SaveLink(ctx, link)
}

func (r *PebbleLinkRepository) GetByID(ctx context.Context, id domain.LinkID) (*domain.Link, error) {
	return r.repo.GetLinkByID(ctx, id)
}

func (r *PebbleLinkRepository) ListByUser(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	return r.repo.ListLinksByUser(ctx, userID)
}

func (r *PebbleLinkRepository) Delete(ctx context.Context, id domain.LinkID) error {
	return r.repo.DeleteLink(ctx, id)
}

func (r *PebbleLinkRepository) Reorder(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error {
	return r.repo.Reorder(ctx, userID, linkIDs)
}

// PostgresLinkRepository wraps PostgresRepository to implement domain.LinkRepository.
type PostgresLinkRepository struct {
	repo *PostgresRepository
}

func NewPostgresLinkRepository(repo *PostgresRepository) *PostgresLinkRepository {
	return &PostgresLinkRepository{repo: repo}
}

func (r *PostgresLinkRepository) Save(ctx context.Context, link *domain.Link) error {
	return r.repo.SaveLink(ctx, link)
}

func (r *PostgresLinkRepository) GetByID(ctx context.Context, id domain.LinkID) (*domain.Link, error) {
	return r.repo.GetLinkByID(ctx, id)
}

func (r *PostgresLinkRepository) ListByUser(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	return r.repo.ListLinksByUser(ctx, userID)
}

func (r *PostgresLinkRepository) Delete(ctx context.Context, id domain.LinkID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresLinkRepository) Reorder(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error {
	return r.repo.Reorder(ctx, userID, linkIDs)
}
