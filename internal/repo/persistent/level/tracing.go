package level

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/level"

type tracedRepo struct{ next repo.LevelRepo }

func newTraced(next repo.LevelRepo) repo.LevelRepo { return &tracedRepo{next: next} }

func startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name, trace.WithAttributes(attrs...))
}

func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

func (r *tracedRepo) ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error) {
	ctx, span := startSpan(ctx, "LevelRepo.ListLevels")
	result, err := r.next.ListLevels(ctx, languageID)
	endSpan(span, err)

	return result, err
}

func (r *tracedRepo) CreateLevel(ctx context.Context, level *entity.Level) error {
	ctx, span := startSpan(ctx, "LevelRepo.CreateLevel", attribute.String("level.code", level.Code))
	err := r.next.CreateLevel(ctx, level)
	endSpan(span, err)

	return err
}

func (r *tracedRepo) UpdateLevel(ctx context.Context, level *entity.Level) error {
	ctx, span := startSpan(ctx, "LevelRepo.UpdateLevel", attribute.Int("level.id", level.ID))
	err := r.next.UpdateLevel(ctx, level)
	endSpan(span, err)

	return err
}

func (r *tracedRepo) DeleteLevel(ctx context.Context, id int) error {
	ctx, span := startSpan(ctx, "LevelRepo.DeleteLevel", attribute.Int("level.id", id))
	err := r.next.DeleteLevel(ctx, id)
	endSpan(span, err)

	return err
}
