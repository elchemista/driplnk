package domain

import "errors"

// ErrNotFound is returned when an entity is not found in the repository.
var ErrNotFound = errors.New("not found")
