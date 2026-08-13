package response

import (
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
)

type FeedChannel struct {
	ID        int    `json:"id" example:"1"`
	Name      string `json:"name" example:"BBC Learning English"`
	AvatarURL string `json:"avatar_url" example:"https://cdn.app.com/channels/bbc.png"`
} // @name response.FeedChannel

type FeedVideo struct {
	ID              int64  `json:"id" example:"1"`
	Title           string `json:"title" example:"Learn English with the News"`
	YouTubeID       string `json:"youtube_id" example:"dQw4w9WgXcQ"`
	ThumbnailURL    string `json:"thumbnail_url" example:"https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"`
	DurationSeconds int    `json:"duration_seconds" example:"300"`
	LevelCode       string `json:"level_code" example:"A1"`
} // @name response.FeedVideo

type FeedSection struct {
	Channel FeedChannel `json:"channel"`
	Videos  []FeedVideo `json:"videos"`
} // @name response.FeedSection

type HomeFeed struct {
	Sections []FeedSection `json:"sections"`
} // @name response.HomeFeed

type ChannelVideos struct {
	Videos     []FeedVideo `json:"videos"`
	Page       int         `json:"page" example:"1"`
	Limit      int         `json:"limit" example:"20"`
	Total      int         `json:"total" example:"42"`
	TotalPages int         `json:"total_pages" example:"3"`
} // @name response.ChannelVideos

type SearchVideoLanguage struct {
	ID   int    `json:"id" example:"1"`
	Code string `json:"code" example:"en"`
	Name string `json:"name" example:"English"`
	Flag string `json:"flag" example:"🇬🇧"`
} // @name response.SearchVideoLanguage

type SearchVideo struct {
	FeedVideo
	Language SearchVideoLanguage `json:"language"`
} // @name response.SearchVideo

type SearchVideos struct {
	Videos     []SearchVideo `json:"videos"`
	Page       int           `json:"page" example:"1"`
	Limit      int           `json:"limit" example:"20"`
	Total      int           `json:"total" example:"42"`
	TotalPages int           `json:"total_pages" example:"3"`
} // @name response.SearchVideos

type HistoryVideo struct {
	FeedVideo
	LastPositionSeconds int       `json:"last_position_seconds" example:"75"`
	LastWatchedAt       time.Time `json:"last_watched_at" example:"2026-08-11T10:30:00Z"`
} // @name response.HistoryVideo

type WatchStatus struct {
	Watched bool          `json:"watched" example:"true"`
	Video   *HistoryVideo `json:"video,omitempty"`
} // @name response.WatchStatus

type SavedVideo struct {
	FeedVideo
	CreatedAt time.Time `json:"created_at" example:"2026-08-11T10:30:00Z"`
} // @name response.SavedVideo

type WatchHistory struct {
	Videos     []HistoryVideo `json:"videos"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"20"`
	Total      int            `json:"total" example:"42"`
	TotalPages int            `json:"total_pages" example:"3"`
} // @name response.WatchHistory

type WatchLater struct {
	Videos     []SavedVideo `json:"videos"`
	Page       int          `json:"page" example:"1"`
	Limit      int          `json:"limit" example:"20"`
	Total      int          `json:"total" example:"42"`
	TotalPages int          `json:"total_pages" example:"3"`
} // @name response.WatchLater

type VideoCaptionItem struct {
	ID                 int64  `json:"id" example:"1"`
	SentenceOrder      int    `json:"sentenceOrder" example:"1"`
	StartTimeMS        int64  `json:"startTimeMs" example:"1000"`
	EndTimeMS          int64  `json:"endTimeMs" example:"3500"`
	Text               string `json:"text" example:"Hello"`
	DictationCompleted bool   `json:"dictation_completed" example:"true"`
}

type OriginalVideoCaptions struct {
	VideoID      int64              `json:"videoId"`
	LanguageID   int                `json:"languageId"`
	LanguageCode string             `json:"languageCode"`
	VideoState   entity.VideoState  `json:"video_state"`
	Items        []VideoCaptionItem `json:"items"`
	Total        int                `json:"total"`
}
