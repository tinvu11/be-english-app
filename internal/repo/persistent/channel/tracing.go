package channel

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/channel"

type tracedRepo struct{ next repo.ChannelRepo }

func newTraced(next repo.ChannelRepo) repo.ChannelRepo { return &tracedRepo{next: next} }

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

func (r *tracedRepo) ListChannels(ctx context.Context) ([]entity.Channel, error) {
	ctx, span := startSpan(ctx, "ChannelRepo.ListChannels")
	result, err := r.next.ListChannels(ctx)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) CreateChannel(ctx context.Context, channel *entity.Channel) error {
	ctx, span := startSpan(ctx, "ChannelRepo.CreateChannel", attribute.String("channel.youtube_id", channel.ChannelYouTubeID))
	err := r.next.CreateChannel(ctx, channel)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) UpdateChannel(ctx context.Context, channel *entity.Channel) error {
	ctx, span := startSpan(ctx, "ChannelRepo.UpdateChannel", attribute.Int("channel.id", channel.ID))
	err := r.next.UpdateChannel(ctx, channel)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) DeleteChannel(ctx context.Context, id int) error {
	ctx, span := startSpan(ctx, "ChannelRepo.DeleteChannel", attribute.Int("channel.id", id))
	err := r.next.DeleteChannel(ctx, id)
	endSpan(span, err)
	return err
}
