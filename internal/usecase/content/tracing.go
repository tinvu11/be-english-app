package content

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const _tracerName = "github.com/evrone/go-clean-template/internal/usecase/content"

type tracedUseCase struct {
	next usecase.Content
}

func newTraced(next usecase.Content) usecase.Content {
	return &tracedUseCase{next: next}
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

// CreateTopic implements usecase.Content.
func (t *tracedUseCase) CreateTopic(ctx context.Context, item entity.Topic) (entity.Topic, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.CreateTopic", attribute.String("topic.name", item.Name))
	res, err := t.next.CreateTopic(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("topic.id", res.ID))
	}
	endSpan(span, err)
	return res, err
}

// GetTopic implements usecase.Content.
func (t *tracedUseCase) GetTopic(ctx context.Context, id int64) (entity.Topic, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.GetTopic", attribute.Int64("topic.id", id))
	res, err := t.next.GetTopic(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListTopics implements usecase.Content.
func (t *tracedUseCase) ListTopics(ctx context.Context, limit, offset int) ([]entity.Topic, int, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.ListTopics",
		attribute.Int("content.limit", limit),
		attribute.Int("content.offset", offset),
	)
	res, total, err := t.next.ListTopics(ctx, limit, offset)
	endSpan(span, err)
	return res, total, err
}

// UpdateTopic implements usecase.Content.
func (t *tracedUseCase) UpdateTopic(ctx context.Context, id int64, item entity.Topic) (entity.Topic, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.UpdateTopic", attribute.Int64("topic.id", id))
	res, err := t.next.UpdateTopic(ctx, id, item)
	endSpan(span, err)
	return res, err
}

// DeleteTopic implements usecase.Content.
func (t *tracedUseCase) DeleteTopic(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentUseCase.DeleteTopic", attribute.Int64("topic.id", id))
	err := t.next.DeleteTopic(ctx, id)
	endSpan(span, err)
	return err
}

// CreateLevel implements usecase.Content.
func (t *tracedUseCase) CreateLevel(ctx context.Context, item entity.Level) (entity.Level, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.CreateLevel", attribute.String("level.name", item.Name))
	res, err := t.next.CreateLevel(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("level.id", res.ID))
	}
	endSpan(span, err)
	return res, err
}

// GetLevel implements usecase.Content.
func (t *tracedUseCase) GetLevel(ctx context.Context, id int64) (entity.Level, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.GetLevel", attribute.Int64("level.id", id))
	res, err := t.next.GetLevel(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListLevels implements usecase.Content.
func (t *tracedUseCase) ListLevels(ctx context.Context, limit, offset int) ([]entity.Level, int, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.ListLevels",
		attribute.Int("content.limit", limit),
		attribute.Int("content.offset", offset),
	)
	res, total, err := t.next.ListLevels(ctx, limit, offset)
	endSpan(span, err)
	return res, total, err
}

// UpdateLevel implements usecase.Content.
func (t *tracedUseCase) UpdateLevel(ctx context.Context, id int64, item entity.Level) (entity.Level, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.UpdateLevel", attribute.Int64("level.id", id))
	res, err := t.next.UpdateLevel(ctx, id, item)
	endSpan(span, err)
	return res, err
}

// DeleteLevel implements usecase.Content.
func (t *tracedUseCase) DeleteLevel(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentUseCase.DeleteLevel", attribute.Int64("level.id", id))
	err := t.next.DeleteLevel(ctx, id)
	endSpan(span, err)
	return err
}

// CreateChannel implements usecase.Content.
func (t *tracedUseCase) CreateChannel(ctx context.Context, item entity.Channel) (entity.Channel, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.CreateChannel", attribute.String("channel.name", item.ChannelName))
	res, err := t.next.CreateChannel(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("channel.id", res.ID))
	}
	endSpan(span, err)
	return res, err
}

// GetChannel implements usecase.Content.
func (t *tracedUseCase) GetChannel(ctx context.Context, id int64) (entity.Channel, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.GetChannel", attribute.Int64("channel.id", id))
	res, err := t.next.GetChannel(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListChannels implements usecase.Content.
func (t *tracedUseCase) ListChannels(ctx context.Context, limit, offset int) ([]entity.Channel, int, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.ListChannels",
		attribute.Int("content.limit", limit),
		attribute.Int("content.offset", offset),
	)
	res, total, err := t.next.ListChannels(ctx, limit, offset)
	endSpan(span, err)
	return res, total, err
}

// UpdateChannel implements usecase.Content.
func (t *tracedUseCase) UpdateChannel(ctx context.Context, id int64, item entity.Channel) (entity.Channel, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.UpdateChannel", attribute.Int64("channel.id", id))
	res, err := t.next.UpdateChannel(ctx, id, item)
	endSpan(span, err)
	return res, err
}

// DeleteChannel implements usecase.Content.
func (t *tracedUseCase) DeleteChannel(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentUseCase.DeleteChannel", attribute.Int64("channel.id", id))
	err := t.next.DeleteChannel(ctx, id)
	endSpan(span, err)
	return err
}

// CreateVideo implements usecase.Content.
func (t *tracedUseCase) CreateVideo(ctx context.Context, item entity.Video) (entity.Video, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.CreateVideo", attribute.String("video.title", item.Title))
	res, err := t.next.CreateVideo(ctx, item)
	if err == nil {
		span.SetAttributes(attribute.Int64("video.id", res.ID))
	}
	endSpan(span, err)
	return res, err
}

// GetVideo implements usecase.Content.
func (t *tracedUseCase) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.GetVideo", attribute.Int64("video.id", id))
	res, err := t.next.GetVideo(ctx, id)
	endSpan(span, err)
	return res, err
}

// ListVideos implements usecase.Content.
func (t *tracedUseCase) ListVideos(ctx context.Context, limit, offset int) ([]entity.Video, int, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.ListVideos",
		attribute.Int("content.limit", limit),
		attribute.Int("content.offset", offset),
	)
	res, total, err := t.next.ListVideos(ctx, limit, offset)
	endSpan(span, err)
	return res, total, err
}

// UpdateVideo implements usecase.Content.
func (t *tracedUseCase) UpdateVideo(ctx context.Context, id int64, item entity.Video) (entity.Video, error) {
	ctx, span := startSpan(ctx, "ContentUseCase.UpdateVideo", attribute.Int64("video.id", id))
	res, err := t.next.UpdateVideo(ctx, id, item)
	endSpan(span, err)
	return res, err
}

// DeleteVideo implements usecase.Content.
func (t *tracedUseCase) DeleteVideo(ctx context.Context, id int64) error {
	ctx, span := startSpan(ctx, "ContentUseCase.DeleteVideo", attribute.Int64("video.id", id))
	err := t.next.DeleteVideo(ctx, id)
	endSpan(span, err)
	return err
}
