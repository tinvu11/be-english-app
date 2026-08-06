package adminuser

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/adminuser"

type tracedRepo struct{ next repo.AdminUserRepo }

func newTraced(next repo.AdminUserRepo) repo.AdminUserRepo { return &tracedRepo{next: next} }
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
func (r *tracedRepo) ListUsers(ctx context.Context, f entity.UserFilter) (entity.UserList, error) {
	ctx, s := startSpan(ctx, "AdminUserRepo.ListUsers")
	v, e := r.next.ListUsers(ctx, f)
	endSpan(s, e)
	return v, e
}
func (r *tracedRepo) SetUserActive(ctx context.Context, id string, a bool) (entity.User, error) {
	ctx, s := startSpan(ctx, "AdminUserRepo.SetUserActive", attribute.String("user.id", id))
	v, e := r.next.SetUserActive(ctx, id, a)
	endSpan(s, e)
	return v, e
}
func (r *tracedRepo) SetUserRole(ctx context.Context, id, role string) (entity.User, error) {
	ctx, s := startSpan(ctx, "AdminUserRepo.SetUserRole", attribute.String("user.id", id))
	v, e := r.next.SetUserRole(ctx, id, role)
	endSpan(s, e)
	return v, e
}
