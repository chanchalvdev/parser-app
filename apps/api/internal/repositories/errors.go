package repositories

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrRepositoryUnready = errors.New("repository is not initialized")
)

