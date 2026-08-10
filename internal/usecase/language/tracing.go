package language

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/language"

type tracedUseCase struct {
	next usecase.Language
}

func newTraced(next usecase.Language) usecase.Language {
	return &tracedUseCase{next: next}
}

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

func (uc *tracedUseCase) ListLanguages(ctx context.Context) ([]entity.Language, error) {
	ctx, span := startSpan(ctx, "LanguageUseCase.ListLanguages")
	result, err := uc.next.ListLanguages(ctx)
	endSpan(span, err)

	return result, err
}

func (uc *tracedUseCase) CreateLanguage(ctx context.Context, code, name, flagEmoji string, isLearnable bool) (entity.Language, error) {
	ctx, span := startSpan(ctx, "LanguageUseCase.CreateLanguage", attribute.String("language.code", code))
	result, err := uc.next.CreateLanguage(ctx, code, name, flagEmoji, isLearnable)
	endSpan(span, err)

	return result, err
}

func (uc *tracedUseCase) UpdateLanguage(ctx context.Context, id int, name, flagEmoji string, isActive, isLearnable bool) (entity.Language, error) {
	ctx, span := startSpan(ctx, "LanguageUseCase.UpdateLanguage", attribute.Int("language.id", id))
	result, err := uc.next.UpdateLanguage(ctx, id, name, flagEmoji, isActive, isLearnable)
	endSpan(span, err)

	return result, err
}
