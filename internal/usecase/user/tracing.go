package user

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const _tracerName = "github.com/evrone/go-clean-template/internal/usecase/user"

// tracedUseCase wraps a User usecase with OpenTelemetry spans, closing the
// gap between transport spans (HTTP/gRPC/AMQP/NATS) and repository spans.
type tracedUseCase struct {
	next usecase.User
}

// newTraced wraps a User usecase with tracing spans.
func newTraced(next usecase.User) usecase.User {
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

func (u *tracedUseCase) Authenticate(ctx context.Context, identity entity.AuthIdentity) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserUseCase.Authenticate", attribute.String("user.firebase_uid", identity.UID))

	result, err := u.next.Authenticate(ctx, identity)
	endSpan(span, err)

	return result, err
}

func (u *tracedUseCase) Register(ctx context.Context, username, email, password string) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserUseCase.Register", attribute.String("user.email", email))

	result, err := u.next.Register(ctx, username, email, password)
	endSpan(span, err)

	return result, err
}

func (u *tracedUseCase) Login(ctx context.Context, email, password string) (string, error) {
	ctx, span := startSpan(ctx, "UserUseCase.Login", attribute.String("user.email", email))

	result, err := u.next.Login(ctx, email, password)
	endSpan(span, err)

	return result, err
}

func (u *tracedUseCase) GetUser(ctx context.Context, userID string) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserUseCase.GetUser", attribute.String("user.id", userID))

	result, err := u.next.GetUser(ctx, userID)
	endSpan(span, err)

	return result, err
}

func (u *tracedUseCase) UpdateLanguages(ctx context.Context, userID string, nativeLanguageID, targetLanguageID int) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserUseCase.UpdateLanguages", attribute.String("user.id", userID))
	result, err := u.next.UpdateLanguages(ctx, userID, nativeLanguageID, targetLanguageID)
	endSpan(span, err)
	return result, err
}

func (u *tracedUseCase) ListWatchHistory(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	ctx, span := startSpan(ctx, "UserUseCase.ListWatchHistory", attribute.String("user.id", userID))
	result, err := u.next.ListWatchHistory(ctx, userID, limit, offset)
	endSpan(span, err)
	return result, err
}

func (u *tracedUseCase) GetWatchHistory(ctx context.Context, userID string, videoID int64) (entity.UserVideo, bool, error) {
	ctx, span := startSpan(ctx, "UserUseCase.GetWatchHistory", attribute.String("user.id", userID), attribute.Int64("video.id", videoID))
	result, watched, err := u.next.GetWatchHistory(ctx, userID, videoID)
	endSpan(span, err)
	return result, watched, err
}

func (u *tracedUseCase) ListWatchLater(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	ctx, span := startSpan(ctx, "UserUseCase.ListWatchLater", attribute.String("user.id", userID))
	result, err := u.next.ListWatchLater(ctx, userID, limit, offset)
	endSpan(span, err)
	return result, err
}

func (u *tracedUseCase) UpsertWatchHistory(ctx context.Context, userID string, videoID int64, lastPositionSeconds int) error {
	ctx, span := startSpan(ctx, "UserUseCase.UpsertWatchHistory", attribute.String("user.id", userID))
	err := u.next.UpsertWatchHistory(ctx, userID, videoID, lastPositionSeconds)
	endSpan(span, err)
	return err
}

func (u *tracedUseCase) RemoveWatchHistory(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserUseCase.RemoveWatchHistory", attribute.String("user.id", userID))
	err := u.next.RemoveWatchHistory(ctx, userID, videoID)
	endSpan(span, err)
	return err
}

func (u *tracedUseCase) SaveWatchLater(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserUseCase.SaveWatchLater", attribute.String("user.id", userID))
	err := u.next.SaveWatchLater(ctx, userID, videoID)
	endSpan(span, err)
	return err
}

func (u *tracedUseCase) RemoveWatchLater(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserUseCase.RemoveWatchLater", attribute.String("user.id", userID))
	err := u.next.RemoveWatchLater(ctx, userID, videoID)
	endSpan(span, err)
	return err
}

func (u *tracedUseCase) CompleteDictation(ctx context.Context, userID string, videoID, captionID int64) (entity.DictationProgress, error) {
	ctx, span := startSpan(ctx, "UserUseCase.CompleteDictation", attribute.String("user.id", userID),
		attribute.Int64("video.id", videoID), attribute.Int64("caption.id", captionID))
	result, err := u.next.CompleteDictation(ctx, userID, videoID, captionID)
	endSpan(span, err)
	return result, err
}

func (u *tracedUseCase) ListCompletedDictations(ctx context.Context, userID string, videoID int64) (entity.DictationProgressList, error) {
	ctx, span := startSpan(ctx, "UserUseCase.ListCompletedDictations", attribute.String("user.id", userID), attribute.Int64("video.id", videoID))
	result, err := u.next.ListCompletedDictations(ctx, userID, videoID)
	endSpan(span, err)
	return result, err
}
