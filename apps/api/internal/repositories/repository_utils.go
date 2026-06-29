package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func validatePool(pool interface{}) error {
	if pool == nil {
		return ErrRepositoryUnready
	}
	return nil
}

func mapNoRows(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func asJSON(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "{}", nil
	}
	if !json.Valid(raw) {
		return "", errors.New("invalid JSON payload")
	}
	return string(raw), nil
}

func scanJSON(raw any) (json.RawMessage, error) {
	switch value := raw.(type) {
	case nil:
		return json.RawMessage("{}"), nil
	case []byte:
		if len(value) == 0 {
			return json.RawMessage("{}"), nil
		}
		if !json.Valid(value) {
			return nil, errors.New("invalid json in database")
		}
		return append(json.RawMessage(nil), value...), nil
	case string:
		if len(value) == 0 {
			return json.RawMessage("{}"), nil
		}
		if !json.Valid([]byte(value)) {
			return nil, errors.New("invalid json in database")
		}
		return json.RawMessage(value), nil
	default:
		return nil, fmt.Errorf("unsupported json value type %T", raw)
	}
}

func nullString(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

func rowsAffectedError(tag pgconn.CommandTag, entity string) error {
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, entity)
	}
	return nil
}
