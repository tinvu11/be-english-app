package topic

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/topic"

type tracedRepo struct{ next repo.TopicRepo }

func newTraced(next repo.TopicRepo) repo.TopicRepo { return &tracedRepo{next: next} }
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
func (r *tracedRepo) ListTopics(ctx context.Context) ([]entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicRepo.ListTopics")
	v, e := r.next.ListTopics(ctx)
	endSpan(s, e)
	return v, e
}
func (r *tracedRepo) CreateTopic(ctx context.Context, v *entity.Topic) error {
	ctx, s := startSpan(ctx, "TopicRepo.CreateTopic", attribute.String("topic.slug", v.Slug))
	e := r.next.CreateTopic(ctx, v)
	endSpan(s, e)
	return e
}
func (r *tracedRepo) UpdateTopic(ctx context.Context, v *entity.Topic) error {
	ctx, s := startSpan(ctx, "TopicRepo.UpdateTopic", attribute.Int("topic.id", v.ID))
	e := r.next.UpdateTopic(ctx, v)
	endSpan(s, e)
	return e
}
func (r *tracedRepo) SetTopicActive(ctx context.Context, id int, a bool) (entity.Topic, error) {
	ctx, s := startSpan(ctx, "TopicRepo.SetTopicActive", attribute.Int("topic.id", id))
	v, e := r.next.SetTopicActive(ctx, id, a)
	endSpan(s, e)
	return v, e
}
func (r *tracedRepo) DeleteTopic(ctx context.Context, id int) error {
	ctx, s := startSpan(ctx, "TopicRepo.DeleteTopic", attribute.Int("topic.id", id))
	e := r.next.DeleteTopic(ctx, id)
	endSpan(s, e)
	return e
}
