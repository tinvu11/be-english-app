package user

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const _tracerName = "github.com/evrone/go-clean-template/internal/repo/persistent/user"

// tracedRepo wraps a UserRepo with OpenTelemetry spans.
type tracedRepo struct {
	next repo.UserRepo
}

func newTraced(next repo.UserRepo) repo.UserRepo {
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

func (r *tracedRepo) Store(ctx context.Context, user *entity.User) error {
	ctx, span := startSpan(ctx, "UserRepo.Store", attribute.String("user.id", user.ID))

	err := r.next.Store(ctx, user)
	endSpan(span, err)

	return err
}

func (r *tracedRepo) GetByID(ctx context.Context, id string) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserRepo.GetByID", attribute.String("user.id", id))

	result, err := r.next.GetByID(ctx, id)
	endSpan(span, err)

	return result, err
}

func (r *tracedRepo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserRepo.GetByEmail", attribute.String("user.email", email))

	result, err := r.next.GetByEmail(ctx, email)
	endSpan(span, err)

	return result, err
}

func (r *tracedRepo) GetByFirebaseUID(ctx context.Context, firebaseUID string) (entity.User, error) {
	ctx, span := startSpan(ctx, "UserRepo.GetByFirebaseUID", attribute.String("user.firebase_uid", firebaseUID))

	result, err := r.next.GetByFirebaseUID(ctx, firebaseUID)
	endSpan(span, err)

	return result, err
}

func (r *tracedRepo) UpdateFirebaseProfile(ctx context.Context, id, username, avatarURL string) error {
	ctx, span := startSpan(ctx, "UserRepo.UpdateFirebaseProfile", attribute.String("user.id", id))

	err := r.next.UpdateFirebaseProfile(ctx, id, username, avatarURL)
	endSpan(span, err)

	return err
}

func (r *tracedRepo) UpdateLanguages(ctx context.Context, id string, nativeLanguageID, targetLanguageID int) error {
	ctx, span := startSpan(ctx, "UserRepo.UpdateLanguages", attribute.String("user.id", id))
	err := r.next.UpdateLanguages(ctx, id, nativeLanguageID, targetLanguageID)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) ListWatchHistory(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	ctx, span := startSpan(ctx, "UserRepo.ListWatchHistory", attribute.String("user.id", userID))
	result, err := r.next.ListWatchHistory(ctx, userID, limit, offset)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) GetWatchHistory(ctx context.Context, userID string, videoID int64) (entity.UserVideo, bool, error) {
	ctx, span := startSpan(ctx, "UserRepo.GetWatchHistory", attribute.String("user.id", userID), attribute.Int64("video.id", videoID))
	result, watched, err := r.next.GetWatchHistory(ctx, userID, videoID)
	endSpan(span, err)
	return result, watched, err
}

func (r *tracedRepo) ListWatchLater(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	ctx, span := startSpan(ctx, "UserRepo.ListWatchLater", attribute.String("user.id", userID))
	result, err := r.next.ListWatchLater(ctx, userID, limit, offset)
	endSpan(span, err)
	return result, err
}

func (r *tracedRepo) UpsertWatchHistory(ctx context.Context, userID string, videoID int64, lastPositionSeconds int) error {
	ctx, span := startSpan(ctx, "UserRepo.UpsertWatchHistory", attribute.String("user.id", userID))
	err := r.next.UpsertWatchHistory(ctx, userID, videoID, lastPositionSeconds)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) RemoveWatchHistory(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserRepo.RemoveWatchHistory", attribute.String("user.id", userID))
	err := r.next.RemoveWatchHistory(ctx, userID, videoID)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) SaveWatchLater(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserRepo.SaveWatchLater", attribute.String("user.id", userID))
	err := r.next.SaveWatchLater(ctx, userID, videoID)
	endSpan(span, err)
	return err
}

func (r *tracedRepo) RemoveWatchLater(ctx context.Context, userID string, videoID int64) error {
	ctx, span := startSpan(ctx, "UserRepo.RemoveWatchLater", attribute.String("user.id", userID))
	err := r.next.RemoveWatchLater(ctx, userID, videoID)
	endSpan(span, err)
	return err
}
