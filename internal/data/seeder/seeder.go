package seeder

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seeds/*.sql
var seedsFS embed.FS

type Seeder struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Seeder {
	return &Seeder{pool: pool}
}

func (s *Seeder) Run(ctx context.Context, env string) error {
	files, err := seedsFS.ReadDir("seeds")
	if err != nil {
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if env != "" && !strings.Contains(file.Name(), env) {
			continue
		}

		data, err := seedsFS.ReadFile(filepath.Join("seeds", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read seed file: %w", err)
		}

		if _, err := s.pool.Exec(ctx, string(data)); err != nil {
			return fmt.Errorf("seed %s failed: %w", file.Name(), err)
		}
	}
	return nil
}
