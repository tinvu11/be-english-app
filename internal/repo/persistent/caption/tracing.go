package caption

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracedRepo struct{ next repo.CaptionRepo }

func newTraced(next repo.CaptionRepo) repo.CaptionRepo { return &tracedRepo{next: next} }

func (r *tracedRepo) ListCaptions(ctx context.Context, videoID int64) ([]entity.Caption, error) {
	ctx, span := otel.Tracer("repo.caption").Start(ctx, "CaptionRepo.ListCaptions")
	span.SetAttributes(attribute.Int64("video.id", videoID))
	items, err := r.next.ListCaptions(ctx, videoID)
	finishSpan(span, err)
	return items, err
}
func (r *tracedRepo) CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error) {
	ctx, span := otel.Tracer("repo.caption").Start(ctx, "CaptionRepo.CreateCaption")
	item, err := r.next.CreateCaption(ctx, videoID, input)
	finishSpan(span, err)
	return item, err
}
func (r *tracedRepo) ImportCaptions(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error) {
	ctx, span := otel.Tracer("repo.caption").Start(ctx, "CaptionRepo.ImportCaptions")
	items, err := r.next.ImportCaptions(ctx, videoID, inputs)
	finishSpan(span, err)
	return items, err
}
func (r *tracedRepo) UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error) {
	ctx, span := otel.Tracer("repo.caption").Start(ctx, "CaptionRepo.UpdateCaption")
	item, err := r.next.UpdateCaption(ctx, videoID, captionID, input)
	finishSpan(span, err)
	return item, err
}
func (r *tracedRepo) DeleteCaption(ctx context.Context, videoID, captionID int64) error {
	ctx, span := otel.Tracer("repo.caption").Start(ctx, "CaptionRepo.DeleteCaption")
	err := r.next.DeleteCaption(ctx, videoID, captionID)
	finishSpan(span, err)
	return err
}

func finishSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
