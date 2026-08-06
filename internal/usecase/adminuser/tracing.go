package adminuser

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/evrone/go-clean-template/internal/usecase/adminuser"

type tracedUseCase struct{ next usecase.AdminUser }

func newTraced(next usecase.AdminUser) usecase.AdminUser { return &tracedUseCase{next: next} }
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
func (u *tracedUseCase) ListUsers(ctx context.Context, f entity.UserFilter) (entity.UserList, error) {
	ctx, s := startSpan(ctx, "AdminUserUseCase.ListUsers")
	v, e := u.next.ListUsers(ctx, f)
	endSpan(s, e)
	return v, e
}
func (u *tracedUseCase) SetUserActive(ctx context.Context, a, id string, v bool) (entity.User, error) {
	ctx, s := startSpan(ctx, "AdminUserUseCase.SetUserActive", attribute.String("user.id", id))
	x, e := u.next.SetUserActive(ctx, a, id, v)
	endSpan(s, e)
	return x, e
}
func (u *tracedUseCase) SetUserRole(ctx context.Context, a, id, role string) (entity.User, error) {
	ctx, s := startSpan(ctx, "AdminUserUseCase.SetUserRole", attribute.String("user.id", id))
	v, e := u.next.SetUserRole(ctx, a, id, role)
	endSpan(s, e)
	return v, e
}
