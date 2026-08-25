// Package quota implements atomic PostgreSQL-backed feature usage counters.
package quota

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) repo.QuotaRepo { return &Repo{pg} }

func (r *Repo) Consume(ctx context.Context, userID string, feature entity.FeatureKey, windowStart time.Time, limit int) (
	used int, allowed bool, err error,
) {
	err = r.Pool.QueryRow(ctx, `INSERT INTO feature_usage_counters(user_id,feature_key,window_start,used_count)
		VALUES($1,$2,$3,1)
		ON CONFLICT(user_id,feature_key,window_start) DO UPDATE
		SET used_count=feature_usage_counters.used_count+1,updated_at=CURRENT_TIMESTAMP
		WHERE feature_usage_counters.used_count < $4
		RETURNING used_count`, userID, feature, windowStart, limit).Scan(&used)
	if errors.Is(err, pgx.ErrNoRows) {
		return limit, false, nil
	}

	if err != nil {
		return 0, false, fmt.Errorf("QuotaRepo - Consume: %w", err)
	}

	return used, true, nil
}
