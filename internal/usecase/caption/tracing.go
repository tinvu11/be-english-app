package caption

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracedUseCase struct{ next usecase.Caption }

func newTraced(next usecase.Caption) usecase.Caption { return &tracedUseCase{next: next} }

func (u *tracedUseCase) ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error) {
	ctx, span := otel.Tracer("usecase.caption").Start(ctx, "CaptionUseCase.ListCaptions")
	items, err := u.next.ListCaptions(ctx, videoID, filter)
	finishSpan(span, err)
	return items, err
}
func (u *tracedUseCase) CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error) {
	ctx, span := otel.Tracer("usecase.caption").Start(ctx, "CaptionUseCase.CreateCaption")
	item, err := u.next.CreateCaption(ctx, videoID, input)
	finishSpan(span, err)
	return item, err
}
func (u *tracedUseCase) ImportSRT(ctx context.Context, videoID int64, original []byte, translations []entity.SRTTranslationFile) ([]entity.Caption, error) {
	ctx, span := otel.Tracer("usecase.caption").Start(ctx, "CaptionUseCase.ImportSRT")
	items, err := u.next.ImportSRT(ctx, videoID, original, translations)
	finishSpan(span, err)
	return items, err
}
func (u *tracedUseCase) UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error) {
	ctx, span := otel.Tracer("usecase.caption").Start(ctx, "CaptionUseCase.UpdateCaption")
	item, err := u.next.UpdateCaption(ctx, videoID, captionID, input)
	finishSpan(span, err)
	return item, err
}
func (u *tracedUseCase) DeleteCaption(ctx context.Context, videoID, captionID int64) error {
	ctx, span := otel.Tracer("usecase.caption").Start(ctx, "CaptionUseCase.DeleteCaption")
	err := u.next.DeleteCaption(ctx, videoID, captionID)
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
