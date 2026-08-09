package language

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/language"

type tracedRepo struct {
	next repo.LanguageRepo
}

func newTraced(next repo.LanguageRepo) repo.LanguageRepo {
	return &tracedRepo{next: next}
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

func (r *tracedRepo) ListLanguages(ctx context.Context) ([]entity.Language, error) {
	ctx, span := startSpan(ctx, "LanguageRepo.ListLanguages")
	result, err := r.next.ListLanguages(ctx)
	endSpan(span, err)

	return result, err
}

func (r *tracedRepo) GetLanguage(ctx context.Context, id int) (entity.Language, error) {
	ctx, span := startSpan(ctx, "LanguageRepo.GetLanguage", attribute.Int("language.id", id))
	result, err := r.next.GetLanguage(ctx, id)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) CreateLanguage(ctx context.Context, language *entity.Language) error {
	ctx, span := startSpan(ctx, "LanguageRepo.CreateLanguage", attribute.String("language.code", language.Code))
	err := r.next.CreateLanguage(ctx, language)
	endSpan(span, err)

	return err
}

func (r *tracedRepo) UpdateLanguage(ctx context.Context, language *entity.Language) error {
	ctx, span := startSpan(ctx, "LanguageRepo.UpdateLanguage", attribute.Int("language.id", language.ID))
	err := r.next.UpdateLanguage(ctx, language)
	endSpan(span, err)

	return err
}
