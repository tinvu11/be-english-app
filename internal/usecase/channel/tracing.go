package channel

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/channel"

type tracedUseCase struct{ next usecase.Channel }

func newTraced(next usecase.Channel) usecase.Channel { return &tracedUseCase{next: next} }
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
func (u *tracedUseCase) ListChannels(ctx context.Context) ([]entity.Channel, error) {
	ctx, s := startSpan(ctx, "ChannelUseCase.ListChannels")
	v, e := u.next.ListChannels(ctx)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) CreateChannel(ctx context.Context, v entity.Channel) (entity.Channel, error) {
	ctx, s := startSpan(ctx, "ChannelUseCase.CreateChannel", attribute.String("channel.youtube_id", v.ChannelYouTubeID))
	x, e := u.next.CreateChannel(ctx, v)
	endSpan(s, e)
	return x, e
}
func (u *tracedUseCase) UpdateChannel(ctx context.Context, id int, v entity.Channel) (entity.Channel, error) {
	ctx, s := startSpan(ctx, "ChannelUseCase.UpdateChannel", attribute.Int("channel.id", id))
	x, e := u.next.UpdateChannel(ctx, id, v)
	endSpan(s, e)
	return x, e
}
func (u *tracedUseCase) DeleteChannel(ctx context.Context, id int) error {
	ctx, s := startSpan(ctx, "ChannelUseCase.DeleteChannel", attribute.Int("channel.id", id))
	e := u.next.DeleteChannel(ctx, id)
	endSpan(s, e)
	return e
}
