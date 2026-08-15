// Package video implements video management business rules.
package video

import (
	"context"
	"net/url"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const maxPageSize = 100

type UseCase struct {
	repo    repo.VideoRepo
	users   repo.UserRepo
	youtube gateway.YouTubeProvider
}

func New(repository repo.VideoRepo, users repo.UserRepo, youtube gateway.YouTubeProvider) usecase.Video {
	return newTraced(&UseCase{repo: repository, users: users, youtube: youtube})
}

func (uc *UseCase) AddUserYouTubeVideo(ctx context.Context, userID, youtubeURLOrID string) (entity.Video, bool, error) {
	if strings.TrimSpace(userID) == "" {
		return entity.Video{}, false, entity.ErrInvalidVideo
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.Video{}, false, err
	}
	if user.TargetLanguageID == nil || *user.TargetLanguageID <= 0 {
		return entity.Video{}, false, entity.ErrTargetLanguageRequired
	}
	preview, err := uc.PreviewYouTubeVideo(ctx, youtubeURLOrID)
	if err != nil {
		return entity.Video{}, false, err
	}
	return uc.repo.UpsertUserVideo(ctx, userID, preview, *user.TargetLanguageID)
}

func (uc *UseCase) ListUserVideos(ctx context.Context, userID string, limit, offset int) (entity.VideoList, error) {
	if strings.TrimSpace(userID) == "" || limit <= 0 || limit > maxPageSize || offset < 0 {
		return entity.VideoList{}, entity.ErrInvalidVideo
	}
	return uc.repo.ListUserVideos(ctx, userID, limit, offset)
}

func (uc *UseCase) RemoveUserVideo(ctx context.Context, userID string, videoID int64) error {
	if strings.TrimSpace(userID) == "" || videoID <= 0 {
		return entity.ErrInvalidVideo
	}
	return uc.repo.RemoveUserVideo(ctx, userID, videoID)
}

func (uc *UseCase) ListVideos(ctx context.Context, filter entity.VideoFilter) (entity.VideoList, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if filter.Limit <= 0 || filter.Limit > maxPageSize || filter.Offset < 0 || len(filter.Search) > 255 ||
		filter.PerChannelLimit < 0 || filter.PerChannelLimit > maxPageSize ||
		(filter.LanguageID != nil && *filter.LanguageID <= 0) || (filter.LevelID != nil && *filter.LevelID <= 0) ||
		(filter.ChannelID != nil && *filter.ChannelID <= 0) || (filter.TopicID != nil && *filter.TopicID <= 0) ||
		(filter.Status != "" && !validStatus(filter.Status)) {
		return entity.VideoList{}, entity.ErrInvalidVideo
	}
	return uc.repo.ListVideos(ctx, filter)
}

func (uc *UseCase) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	if id <= 0 {
		return entity.Video{}, entity.ErrInvalidVideo
	}
	return uc.repo.GetVideo(ctx, id)
}

func (uc *UseCase) PreviewYouTubeVideo(ctx context.Context, youtubeURLOrID string) (entity.YouTubeVideoPreview, error) {
	youtubeID, err := parseYouTubeID(youtubeURLOrID)
	if err != nil {
		return entity.YouTubeVideoPreview{}, err
	}
	return uc.youtube.PreviewVideo(ctx, youtubeID)
}

func (uc *UseCase) CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error) {
	if err := normalizeAndValidate(&input); err != nil {
		return entity.Video{}, err
	}
	return uc.repo.CreateVideo(ctx, input)
}

func (uc *UseCase) UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error) {
	if id <= 0 {
		return entity.Video{}, entity.ErrInvalidVideo
	}
	if err := normalizeAndValidate(&input); err != nil {
		return entity.Video{}, err
	}
	return uc.repo.UpdateVideo(ctx, id, input)
}

func (uc *UseCase) TransitionVideoStatus(ctx context.Context, id int64, status string) (entity.Video, error) {
	if id <= 0 {
		return entity.Video{}, entity.ErrInvalidVideo
	}
	status = strings.ToLower(strings.TrimSpace(status))
	current, err := uc.repo.GetVideo(ctx, id)
	if err != nil {
		return entity.Video{}, err
	}
	if !validTransition(current.Status, status) {
		return entity.Video{}, entity.ErrInvalidVideoTransition
	}
	return uc.repo.SetVideoStatus(ctx, id, current.Status, status)
}

func (uc *UseCase) DeleteVideo(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidVideo
	}
	return uc.repo.DeleteVideo(ctx, id)
}

func normalizeAndValidate(input *entity.VideoInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.YouTubeID = strings.TrimSpace(input.YouTubeID)
	input.ThumbnailURL = strings.TrimSpace(input.ThumbnailURL)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Title == "" || len(input.Title) > 255 || input.YouTubeID == "" || len(input.YouTubeID) > 50 ||
		input.DurationSeconds < 0 || !validStatus(input.Status) || input.LanguageID <= 0 || input.LevelID <= 0 || input.ChannelID <= 0 {
		return entity.ErrInvalidVideo
	}
	seenTopics := make(map[int]struct{}, len(input.TopicIDs))
	for _, topicID := range input.TopicIDs {
		if topicID <= 0 {
			return entity.ErrInvalidVideo
		}
		if _, exists := seenTopics[topicID]; exists {
			return entity.ErrInvalidVideo
		}
		seenTopics[topicID] = struct{}{}
	}
	return nil
}

func validStatus(status string) bool {
	return status == entity.VideoStatusDraft || status == entity.VideoStatusPublished || status == entity.VideoStatusArchived
}

func validTransition(current, next string) bool {
	return (current == entity.VideoStatusDraft && next == entity.VideoStatusPublished) ||
		(current == entity.VideoStatusPublished && next == entity.VideoStatusArchived)
}

func parseYouTubeID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if validYouTubeID(raw) {
		return raw, nil
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", entity.ErrInvalidVideo
	}
	host := strings.ToLower(parsed.Hostname())
	var id string
	switch host {
	case "youtu.be":
		id = strings.Trim(strings.Split(strings.Trim(parsed.Path, "/"), "/")[0], " ")
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com":
		if parsed.Path == "/watch" {
			id = parsed.Query().Get("v")
		} else {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
				id = parts[1]
			}
		}
	}
	if !validYouTubeID(id) {
		return "", entity.ErrInvalidVideo
	}
	return id, nil
}

func validYouTubeID(value string) bool {
	if len(value) < 6 || len(value) > 50 {
		return false
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}
