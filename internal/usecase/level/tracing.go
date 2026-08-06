package level

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/level"

type tracedUseCase struct{ next usecase.Level }

func newTraced(next usecase.Level) usecase.Level { return &tracedUseCase{next: next} }

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

func (uc *tracedUseCase) ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error) {
	ctx, span := startSpan(ctx, "LevelUseCase.ListLevels")
	result, err := uc.next.ListLevels(ctx, languageID)
	endSpan(span, err)

	return result, err
}

func (uc *tracedUseCase) CreateLevel(ctx context.Context, level entity.Level) (entity.Level, error) {
	ctx, span := startSpan(ctx, "LevelUseCase.CreateLevel", attribute.String("level.code", level.Code))
	result, err := uc.next.CreateLevel(ctx, level)
	endSpan(span, err)

	return result, err
}

func (uc *tracedUseCase) UpdateLevel(ctx context.Context, id int, level entity.Level) (entity.Level, error) {
	ctx, span := startSpan(ctx, "LevelUseCase.UpdateLevel", attribute.Int("level.id", id))
	result, err := uc.next.UpdateLevel(ctx, id, level)
	endSpan(span, err)

	return result, err
}

func (uc *tracedUseCase) DeleteLevel(ctx context.Context, id int) error {
	ctx, span := startSpan(ctx, "LevelUseCase.DeleteLevel", attribute.Int("level.id", id))
	err := uc.next.DeleteLevel(ctx, id)
	endSpan(span, err)

	return err
}
