package content

import (
	"context"
	"math"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const _tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/content"

type tracedRepo struct {
	next repo.ContentRepo
}

func newTraced(next repo.ContentRepo) repo.ContentRepo {
	return &tracedRepo{next: next}
}

func startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer(_tracerName).Start(ctx, name, trace.WithAttributes(attrs...))
}

func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}

func safeUint64ToInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}

	return int64(v)
}

// CreateTopic implements repo.ContentRepo.
func (r *tracedRepo) CreateTopic(ctx context.Context, item *entity.Topic) error {
	ctx, span := startSpan(ctx, "ContentRepo.CreateTopic", attribute.String("topic.name", item.Name))
	err := r.next.CreateTopic(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("topic.id", item.ID))
	}
	endSpan(span, err)
	return err
}

// GetTopic implements repo.ContentRepo.
func (r *tracedRepo) GetTopic(ctx context.Context, id int64) (entity.Topic, error) {
	ctx, span := startSpan(ctx, "ContentRepo.GetTopic", attribute.Int64("topic.id", id))
	res, err := r.next.GetTopic(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListTopics implements repo.ContentRepo.
func (r *tracedRepo) ListTopics(ctx context.Context, f repo.ContentFilter) ([]entity.Topic, int, error) {
	ctx, span := startSpan(ctx, "ContentRepo.ListTopics",
		attribute.Int64("content.limit", safeUint64ToInt64(f.Limit)),
		attribute.Int64("content.offset", safeUint64ToInt64(f.Offset)),
	)
	res, total, err := r.next.ListTopics(ctx, f)
	endSpan(span, err)
	return res, total, err
}

// UpdateTopic implements repo.ContentRepo.
func (r *tracedRepo) UpdateTopic(ctx context.Context, item *entity.Topic) error {
	ctx, span := startSpan(ctx, "ContentRepo.UpdateTopic", attribute.Int64("topic.id", item.ID))
	err := r.next.UpdateTopic(ctx, item)
	endSpan(span, err)
	return err
}

// DeleteTopic implements repo.ContentRepo.
func (r *tracedRepo) DeleteTopic(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentRepo.DeleteTopic", attribute.Int64("topic.id", id))
	err := r.next.DeleteTopic(ctx, id)
	endSpan(span, err)
	return err
}

// CreateLevel implements repo.ContentRepo.
func (r *tracedRepo) CreateLevel(ctx context.Context, item *entity.Level) error {
	ctx, span := startSpan(ctx, "ContentRepo.CreateLevel", attribute.String("level.name", item.Name))
	err := r.next.CreateLevel(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("level.id", item.ID))
	}
	endSpan(span, err)
	return err
}

// GetLevel implements repo.ContentRepo.
func (r *tracedRepo) GetLevel(ctx context.Context, id int64) (entity.Level, error) {
	ctx, span := startSpan(ctx, "ContentRepo.GetLevel", attribute.Int64("level.id", id))
	res, err := r.next.GetLevel(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListLevels implements repo.ContentRepo.
func (r *tracedRepo) ListLevels(ctx context.Context, f repo.ContentFilter) ([]entity.Level, int, error) {
	ctx, span := startSpan(ctx, "ContentRepo.ListLevels",
		attribute.Int64("content.limit", safeUint64ToInt64(f.Limit)),
		attribute.Int64("content.offset", safeUint64ToInt64(f.Offset)),
	)
	res, total, err := r.next.ListLevels(ctx, f)
	endSpan(span, err)
	return res, total, err
}

// UpdateLevel implements repo.ContentRepo.
func (r *tracedRepo) UpdateLevel(ctx context.Context, item *entity.Level) error {
	ctx, span := startSpan(ctx, "ContentRepo.UpdateLevel", attribute.Int64("level.id", item.ID))
	err := r.next.UpdateLevel(ctx, item)
	endSpan(span, err)
	return err
}

// DeleteLevel implements repo.ContentRepo.
func (r *tracedRepo) DeleteLevel(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentRepo.DeleteLevel", attribute.Int64("level.id", id))
	err := r.next.DeleteLevel(ctx, id)
	endSpan(span, err)
	return err
}

// CreateChannel implements repo.ContentRepo.
func (r *tracedRepo) CreateChannel(ctx context.Context, item *entity.Channel) error {
	ctx, span := startSpan(ctx, "ContentRepo.CreateChannel", attribute.String("channel.name", item.ChannelName))
	err := r.next.CreateChannel(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("channel.id", item.ID))
	}
	endSpan(span, err)
	return err
}

// GetChannel implements repo.ContentRepo.
func (r *tracedRepo) GetChannel(ctx context.Context, id int64) (entity.Channel, error) {
	ctx, span := startSpan(ctx, "ContentRepo.GetChannel", attribute.Int64("channel.id", id))
	res, err := r.next.GetChannel(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListChannels implements repo.ContentRepo.
func (r *tracedRepo) ListChannels(ctx context.Context, f repo.ContentFilter) ([]entity.Channel, int, error) {
	ctx, span := startSpan(ctx, "ContentRepo.ListChannels",
		attribute.Int64("content.limit", safeUint64ToInt64(f.Limit)),
		attribute.Int64("content.offset", safeUint64ToInt64(f.Offset)),
	)
	res, total, err := r.next.ListChannels(ctx, f)
	endSpan(span, err)
	return res, total, err
}

// UpdateChannel implements repo.ContentRepo.
func (r *tracedRepo) UpdateChannel(ctx context.Context, item *entity.Channel) error {
	ctx, span := startSpan(ctx, "ContentRepo.UpdateChannel", attribute.Int64("channel.id", item.ID))
	err := r.next.UpdateChannel(ctx, item)
	endSpan(span, err)
	return err
}

// DeleteChannel implements repo.ContentRepo.
func (r *tracedRepo) DeleteChannel(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentRepo.DeleteChannel", attribute.Int64("channel.id", id))
	err := r.next.DeleteChannel(ctx, id)
	endSpan(span, err)
	return err
}

// CreateVideo implements repo.ContentRepo.
func (r *tracedRepo) CreateVideo(ctx context.Context, item *entity.Video) error {
	ctx, span := startSpan(ctx, "ContentRepo.CreateVideo", attribute.String("video.title", item.Title))
	err := r.next.CreateVideo(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("video.id", item.ID))
	}
	endSpan(span, err)
	return err
}

// GetVideo implements repo.ContentRepo.
func (r *tracedRepo) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	ctx, span := startSpan(ctx, "ContentRepo.GetVideo", attribute.Int64("video.id", id))
	res, err := r.next.GetVideo(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListVideos implements repo.ContentRepo.
func (r *tracedRepo) ListVideos(ctx context.Context, f repo.ContentFilter) ([]entity.Video, int, error) {
	ctx, span := startSpan(ctx, "ContentRepo.ListVideos",
		attribute.Int64("content.limit", safeUint64ToInt64(f.Limit)),
		attribute.Int64("content.offset", safeUint64ToInt64(f.Offset)),
	)
	res, total, err := r.next.ListVideos(ctx, f)
	endSpan(span, err)
	return res, total, err
}

// UpdateVideo implements repo.ContentRepo.
func (r *tracedRepo) UpdateVideo(ctx context.Context, item *entity.Video) error {
	ctx, span := startSpan(ctx, "ContentRepo.UpdateVideo", attribute.Int64("video.id", item.ID))
	err := r.next.UpdateVideo(ctx, item)
	endSpan(span, err)
	return err
}

// DeleteVideo implements repo.ContentRepo.
func (r *tracedRepo) DeleteVideo(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentRepo.DeleteVideo", attribute.Int64("video.id", id))
	err := r.next.DeleteVideo(ctx, id)
	endSpan(span, err)
	return err
}
