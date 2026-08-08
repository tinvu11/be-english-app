package video

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/video"

type tracedUseCase struct{ next usecase.Video }

func newTraced(next usecase.Video) usecase.Video { return &tracedUseCase{next: next} }
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
func (u *tracedUseCase) ListVideos(ctx context.Context, f entity.VideoFilter) (entity.VideoList, error) {
	ctx, s := startSpan(ctx, "VideoUseCase.ListVideos")
	v, e := u.next.ListVideos(ctx, f)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	ctx, s := startSpan(ctx, "VideoUseCase.GetVideo", attribute.Int64("video.id", id))
	v, e := u.next.GetVideo(ctx, id)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) PreviewYouTubeVideo(ctx context.Context, youtubeURLOrID string) (entity.YouTubeVideoPreview, error) {
	ctx, span := startSpan(ctx, "VideoUseCase.PreviewYouTubeVideo")
	item, err := u.next.PreviewYouTubeVideo(ctx, youtubeURLOrID)
	endSpan(span, err)
	return item, err
}
func (u *tracedUseCase) CreateVideo(ctx context.Context, i entity.VideoInput) (entity.Video, error) {
	ctx, s := startSpan(ctx, "VideoUseCase.CreateVideo")
	v, e := u.next.CreateVideo(ctx, i)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) UpdateVideo(ctx context.Context, id int64, i entity.VideoInput) (entity.Video, error) {
	ctx, s := startSpan(ctx, "VideoUseCase.UpdateVideo", attribute.Int64("video.id", id))
	v, e := u.next.UpdateVideo(ctx, id, i)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) TransitionVideoStatus(ctx context.Context, id int64, status string) (entity.Video, error) {
	ctx, s := startSpan(ctx, "VideoUseCase.TransitionVideoStatus", attribute.Int64("video.id", id))
	v, e := u.next.TransitionVideoStatus(ctx, id, status)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) DeleteVideo(ctx context.Context, id int64) error {
	ctx, s := startSpan(ctx, "VideoUseCase.DeleteVideo", attribute.Int64("video.id", id))
	e := u.next.DeleteVideo(ctx, id)
	endSpan(s, e)
	return e
}
