package video

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/video"

type tracedRepo struct{ next repo.VideoRepo }

func newTraced(next repo.VideoRepo) repo.VideoRepo { return &tracedRepo{next: next} }

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

func (r *tracedRepo) ListVideos(ctx context.Context, filter entity.VideoFilter) (entity.VideoList, error) {
	ctx, span := startSpan(ctx, "VideoRepo.ListVideos")
	result, err := r.next.ListVideos(ctx, filter)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	ctx, span := startSpan(ctx, "VideoRepo.GetVideo", attribute.Int64("video.id", id))
	result, err := r.next.GetVideo(ctx, id)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error) {
	ctx, span := startSpan(ctx, "VideoRepo.CreateVideo", attribute.String("video.youtube_id", input.YouTubeID))
	result, err := r.next.CreateVideo(ctx, input)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error) {
	ctx, span := startSpan(ctx, "VideoRepo.UpdateVideo", attribute.Int64("video.id", id))
	result, err := r.next.UpdateVideo(ctx, id, input)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) SetVideoStatus(ctx context.Context, id int64, expectedStatus, nextStatus string) (entity.Video, error) {
	ctx, span := startSpan(ctx, "VideoRepo.SetVideoStatus", attribute.Int64("video.id", id))
	result, err := r.next.SetVideoStatus(ctx, id, expectedStatus, nextStatus)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) DeleteVideo(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "VideoRepo.DeleteVideo", attribute.Int64("video.id", id))
	err := r.next.DeleteVideo(ctx, id)
	endSpan(span, err)
	return err
}
