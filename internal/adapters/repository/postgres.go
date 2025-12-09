package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/elchemista/driplnk/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg *PostgresConfig) (*PostgresRepository, error) {
	log.Printf("[INFO] Initializing Postgres connection to %s (truncated if sensitive)", cfg.URL) // TODO: Mask password
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres db: %w", err)
	}

	// Configure pooling values
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	repo := &PostgresRepository{db: db}

	// Apply migrations using DSN
	if err := ApplyMigrations(cfg.URL); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("[INFO] Postgres Adapter initialized successfully")
	return repo, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// --- User Repository ---

func (r *PostgresRepository) Save(ctx context.Context, user *domain.User) error {
	seoMetaBytes, err := json.Marshal(user.SEOMeta)
	if err != nil {
		return fmt.Errorf("marshal seo_meta: %w", err)
	}
	themeBytes, err := json.Marshal(user.Theme)
	if err != nil {
		return fmt.Errorf("marshal theme: %w", err)
	}

	query := `
		INSERT INTO users (id, email, handle, title, description, avatar_url, seo_meta, theme, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			handle = EXCLUDED.handle,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			avatar_url = EXCLUDED.avatar_url,
			seo_meta = EXCLUDED.seo_meta,
			theme = EXCLUDED.theme,
			updated_at = EXCLUDED.updated_at;
	`

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Handle,
		user.Title,
		user.Description,
		user.AvatarURL,
		seoMetaBytes,
		themeBytes,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return r.getUserByField(ctx, "id", id)
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getUserByField(ctx, "email", email)
}

func (r *PostgresRepository) GetByHandle(ctx context.Context, handle string) (*domain.User, error) {
	return r.getUserByField(ctx, "handle", handle)
}

func (r *PostgresRepository) getUserByField(ctx context.Context, field string, value interface{}) (*domain.User, error) {
	query := fmt.Sprintf(`
		SELECT id, email, handle, title, description, avatar_url, seo_meta, theme, created_at, updated_at
		FROM users WHERE %s = $1`, field)

	row := r.db.QueryRowContext(ctx, query, value)

	var user domain.User
	var seoMetaBytes, themeBytes []byte

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Handle,
		&user.Title,
		&user.Description,
		&user.AvatarURL,
		&seoMetaBytes,
		&themeBytes,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(seoMetaBytes) > 0 {
		if err := json.Unmarshal(seoMetaBytes, &user.SEOMeta); err != nil {
			return nil, fmt.Errorf("unmarshal seo_meta: %w", err)
		}
	}
	if len(themeBytes) > 0 {
		if err := json.Unmarshal(themeBytes, &user.Theme); err != nil {
			return nil, fmt.Errorf("unmarshal theme: %w", err)
		}
	}

	return &user, nil
}

func (r *PostgresRepository) ListAll(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, email, handle, title, description, avatar_url, seo_meta, theme, created_at, updated_at FROM users ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var seoMetaBytes, themeBytes []byte
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Handle,
			&user.Title,
			&user.Description,
			&user.AvatarURL,
			&seoMetaBytes,
			&themeBytes,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(seoMetaBytes) > 0 {
			if err := json.Unmarshal(seoMetaBytes, &user.SEOMeta); err != nil {
				return nil, fmt.Errorf("unmarshal seo_meta: %w", err)
			}
		}
		if len(themeBytes) > 0 {
			if err := json.Unmarshal(themeBytes, &user.Theme); err != nil {
				return nil, fmt.Errorf("unmarshal theme: %w", err)
			}
		}
		users = append(users, &user)
	}

	return users, nil
}

// --- Link Repository ---

func (r *PostgresRepository) SaveLink(ctx context.Context, link *domain.Link) error {
	metadataBytes, err := json.Marshal(link.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO links (id, user_id, title, url, type, link_order, is_active, metadata, click_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			type = EXCLUDED.type,
			link_order = EXCLUDED.link_order,
			is_active = EXCLUDED.is_active,
			metadata = EXCLUDED.metadata,
			click_count = EXCLUDED.click_count,
			updated_at = EXCLUDED.updated_at;
	`

	_, err = r.db.ExecContext(ctx, query,
		link.ID,
		link.UserID,
		link.Title,
		link.URL,
		link.Type,
		link.Order,
		link.IsActive,
		metadataBytes,
		link.ClickCount,
		link.CreatedAt,
		link.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetLinkByID(ctx context.Context, id domain.LinkID) (*domain.Link, error) {
	query := `
		SELECT id, user_id, title, url, type, link_order, is_active, metadata, click_count, created_at, updated_at
		FROM links WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var link domain.Link
	var metadataBytes []byte

	err := row.Scan(
		&link.ID,
		&link.UserID,
		&link.Title,
		&link.URL,
		&link.Type,
		&link.Order,
		&link.IsActive,
		&metadataBytes,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &link.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &link, nil
}

func (r *PostgresRepository) ListLinksByUser(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	query := `
		SELECT id, user_id, title, url, type, link_order, is_active, metadata, click_count, created_at, updated_at
		FROM links WHERE user_id = $1
		ORDER BY link_order ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*domain.Link
	for rows.Next() {
		var link domain.Link
		var metadataBytes []byte
		if err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.Title,
			&link.URL,
			&link.Type,
			&link.Order,
			&link.IsActive,
			&metadataBytes,
			&link.ClickCount,
			&link.CreatedAt,
			&link.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &link.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		links = append(links, &link)
	}

	return links, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id domain.LinkID) error {
	query := `DELETE FROM links WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) Reorder(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range linkIDs {
		// Ensure the link belongs to the user to avoid cross-polluting
		// Ideally we should check ownership
		query := `UPDATE links SET link_order = $1 WHERE id = $2 AND user_id = $3`
		res, err := tx.ExecContext(ctx, query, i, id, userID)
		if err != nil {
			return err
		}
		rowsHub, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsHub == 0 {
			// Might want to log specific ID not found or not owned
			continue
		}
	}

	return tx.Commit()
}
