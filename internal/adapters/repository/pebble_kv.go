package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/elchemista/driplnk/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
)

type PebbleRepository struct {
	db *pebble.DB
}

func NewPebbleRepository(cfg *PebbleConfig) (*PebbleRepository, error) {
	log.Printf("[INFO] Initializing PebbleDB at %s", cfg.Path)
	opts := &pebble.Options{}
	// Use a subdirectory for the db to avoid clutter
	dbPath := filepath.Join(cfg.Path, "driplnk.db")
	db, err := pebble.Open(dbPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open pebble db: %w", err)
	}
	log.Println("[INFO] PebbleDB Adapter initialized successfully")
	return &PebbleRepository{db: db}, nil
}

func (r *PebbleRepository) Close() error {
	return r.db.Close()
}

// --- User Repository ---

func (r *PebbleRepository) Save(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	batch := r.db.NewBatch()
	defer batch.Close()

	// Main record
	key := []byte(fmt.Sprintf("user:%s", user.ID))
	if err := batch.Set(key, data, pebble.Sync); err != nil {
		return err
	}

	// Email index
	emailKey := []byte(fmt.Sprintf("user:email:%s", user.Email))
	if err := batch.Set(emailKey, []byte(user.ID), pebble.Sync); err != nil {
		return err
	}

	// Handle index
	handleKey := []byte(fmt.Sprintf("user:handle:%s", user.Handle))
	if err := batch.Set(handleKey, []byte(user.ID), pebble.Sync); err != nil {
		return err
	}

	return batch.Commit(pebble.Sync)
}

func (r *PebbleRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	key := []byte(fmt.Sprintf("user:%s", id))
	val, closer, err := r.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var user domain.User
	if err := json.Unmarshal(val, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PebbleRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	emailKey := []byte(fmt.Sprintf("user:email:%s", email))
	val, closer, err := r.db.Get(emailKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	userID := domain.UserID(val)
	return r.GetByID(ctx, userID)
}

func (r *PebbleRepository) GetByHandle(ctx context.Context, handle string) (*domain.User, error) {
	handleKey := []byte(fmt.Sprintf("user:handle:%s", handle))
	val, closer, err := r.db.Get(handleKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	userID := domain.UserID(val)
	return r.GetByID(ctx, userID)
}

func (r *PebbleRepository) ListAll(ctx context.Context) ([]*domain.User, error) {
	// Scan all keys with "user:" prefix (but not "user:email:" or "user:handle:")
	prefix := []byte("user:")
	iter, _ := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
	})
	defer iter.Close()

	var users []*domain.User

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		// Skip index keys (email and handle)
		if len(key) > 10 && (key[5:10] == "email" || key[5:11] == "handle") {
			continue
		}
		// Skip if not a direct user key (user:<id>)
		parts := splitKey(key)
		if len(parts) != 2 || parts[0] != "user" {
			continue
		}

		var user domain.User
		if err := json.Unmarshal(iter.Value(), &user); err != nil {
			continue // Skip malformed entries
		}
		users = append(users, &user)
	}

	return users, nil
}

// --- Link Repository ---

func (r *PebbleRepository) SaveLink(ctx context.Context, link *domain.Link) error {
	data, err := json.Marshal(link)
	if err != nil {
		return err
	}

	batch := r.db.NewBatch()
	defer batch.Close()

	// Main record
	key := []byte(fmt.Sprintf("link:%s", link.ID))
	if err := batch.Set(key, data, pebble.Sync); err != nil {
		return err
	}

	// User links index (for sorting/listing)
	// Key: user:links:<userID>:<order>:<linkID>
	// Note: Reordering might require deleting old index keys. For MV implementation, strict ordering management might be tricky with KV.
	// Simpler index: user:links:<userID>:<linkID> -> empty
	indexKey := []byte(fmt.Sprintf("user:links:%s:%s", link.UserID, link.ID))
	if err := batch.Set(indexKey, []byte{}, pebble.Sync); err != nil {
		return err
	}

	return batch.Commit(pebble.Sync)
}

func (r *PebbleRepository) GetLinkByID(ctx context.Context, id domain.LinkID) (*domain.Link, error) {
	key := []byte(fmt.Sprintf("link:%s", id))
	val, closer, err := r.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var link domain.Link
	if err := json.Unmarshal(val, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *PebbleRepository) ListLinksByUser(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	prefix := []byte(fmt.Sprintf("user:links:%s:", userID))
	iter, _ := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
	})
	defer iter.Close()

	var links []*domain.Link

	for iter.SeekGE(prefix); iter.Valid() && string(iter.Key())[0:len(prefix)] == string(prefix); iter.Next() {
		// Extract LinkID from key or load it
		// The key is user:links:<userID>:<linkID>
		// We can parse the LinkID from the key
		var linkIDStr string
		fmt.Sscanf(string(iter.Key()), "user:links:%s:%s", new(string), &linkIDStr)

		// This scanf might be flaky if IDs have colons, better to split string
		// Manual split:
		parts := splitKey(string(iter.Key()))
		if len(parts) < 4 {
			continue
		}
		linkID := domain.LinkID(parts[3])

		link, err := r.GetLinkByID(ctx, linkID)
		if err == nil {
			links = append(links, link)
		}
	}
	return links, nil
}

func (r *PebbleRepository) DeleteLink(ctx context.Context, id domain.LinkID) error {
	link, err := r.GetLinkByID(ctx, id)
	if err != nil {
		return err // Or return nil if already gone
	}

	batch := r.db.NewBatch()
	defer batch.Close()

	// Delete Main record
	key := []byte(fmt.Sprintf("link:%s", id))
	batch.Delete(key, pebble.Sync)

	// Delete Index
	indexKey := []byte(fmt.Sprintf("user:links:%s:%s", link.UserID, id))
	batch.Delete(indexKey, pebble.Sync)

	return batch.Commit(pebble.Sync)
}

// Helper to split keys
func splitKey(key string) []string {
	// Simple implementation
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	parts = append(parts, key[start:])
	return parts
}

// Reorder is complex in KV, skipping for initial scaffolding
func (r *PebbleRepository) Reorder(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error {
	return nil
}
