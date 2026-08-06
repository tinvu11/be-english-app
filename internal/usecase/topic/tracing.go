package topic

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/topic"

type tracedUseCase struct{ next usecase.Topic }

func newTraced(next usecase.Topic) usecase.Topic { return &tracedUseCase{next: next} }
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
func (u *tracedUseCase) ListTopics(ctx context.Context) ([]entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicUseCase.ListTopics")
	v, e := u.next.ListTopics(ctx)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) CreateTopic(ctx context.Context, v entity.Topic) (entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicUseCase.CreateTopic", attribute.String("topic.slug", v.Slug))
	x, e := u.next.CreateTopic(ctx, v)
	endSpan(s, e)
	return x, e
}
func (u *tracedUseCase) UpdateTopic(ctx context.Context, id int, v entity.Topic) (entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicUseCase.UpdateTopic", attribute.Int("topic.id", id))
	x, e := u.next.UpdateTopic(ctx, id, v)
	endSpan(s, e)
	return x, e
}
func (u *tracedUseCase) SetTopicActive(ctx context.Context, id int, a bool) (entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicUseCase.SetTopicActive", attribute.Int("topic.id", id))
	v, e := u.next.SetTopicActive(ctx, id, a)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) DeleteTopic(ctx context.Context, id int) error {
	ctx, s := startSpan(ctx, "TopicUseCase.DeleteTopic", attribute.Int("topic.id", id))
	e := u.next.DeleteTopic(ctx, id)
	endSpan(s, e)
	return e
}
