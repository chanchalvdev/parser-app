package db

import (
\t"context"
\t"fmt"
\t"os"
\t"path/filepath"
\t"sort"
\t"strings"

\t"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
\tpool, err := pgxpool.New(ctx, dsn)
\tif err != nil {
\t\treturn nil, fmt.Errorf("create pool: %w", err)
\t}
\tif err := pool.Ping(ctx); err != nil {
\t\treturn nil, err
\t}
\treturn pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
\tentries, err := os.ReadDir(dir)
\tif err != nil {
\t\tif os.IsNotExist(err) {
\t\t\treturn nil
\t\t}
\t\treturn err
\t}
\tsqlFiles := make([]string, 0)
\tfor _, entry := range entries {
\t\tif entry.IsDir() {
\t\t\tcontinue
\t\t}
\t\tname := entry.Name()
\t\tif strings.HasSuffix(name, ".sql") {
\t\t\tsqlFiles = append(sqlFiles, name)
\t\t}
\t}
\tsort.Strings(sqlFiles)
\tfor _, name := range sqlFiles {
\t\tp := filepath.Join(dir, name)
\t\tb, err := os.ReadFile(p)
\t\tif err != nil {
\t\t\treturn fmt.Errorf("read migration %s: %w", p, err)
\t\t}
\t\tif _, err := pool.Exec(ctx, string(b)); err != nil {
\t\t\treturn fmt.Errorf("run migration %s: %w", name, err)
\t\t}
\t}
\treturn nil
}
