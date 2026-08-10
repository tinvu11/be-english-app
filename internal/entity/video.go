package entity

import "time"

const (
	VideoStatusDraft     = "draft"
	VideoStatusPublished = "published"
	VideoStatusArchived  = "archived"
)

type VideoFilter struct {
	LanguageID             *int
	LevelID                *int
	ChannelID              *int
	TopicID                *int
	Status                 string
	Search                 string
	OnlyActiveChannels     bool
	OnlyLearnableLanguages bool
	PerChannelLimit        int
	Limit                  int
	Offset                 int
}

type VideoLanguage struct {
	ID        int    `json:"id" example:"1"`
	Code      string `json:"code" example:"en"`
	Name      string `json:"name" example:"English"`
	FlagEmoji string `json:"flagEmoji" example:"🇬🇧"`
} // @name entity.VideoLanguage

type VideoLevel struct {
	ID   int    `json:"id" example:"2"`
	Code string `json:"code" example:"A2"`
	Name string `json:"name" example:"Elementary"`
} // @name entity.VideoLevel

type VideoChannel struct {
	ID               int    `json:"id" example:"3"`
	ChannelYouTubeID string `json:"channelYoutubeId"`
	Name             string `json:"name" example:"Google Developers"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
} // @name entity.VideoChannel

type VideoTopic struct {
	ID      int    `json:"id" example:"5"`
	Slug    string `json:"slug" example:"music"`
	Name    string `json:"name" example:"Music"`
	IconURL string `json:"iconUrl,omitempty"`
} // @name entity.VideoTopic

type VideoCaptionTranslationAvailability struct {
	LanguageID             int    `json:"languageId" example:"2"`
	LanguageCode           string `json:"languageCode" example:"vi"`
	LanguageName           string `json:"languageName" example:"Vietnamese"`
	TranslatedCaptionCount int    `json:"translatedCaptionCount" example:"120"`
	IsComplete             bool   `json:"isComplete" example:"true"`
} // @name entity.VideoCaptionTranslationAvailability

type VideoCaptionAvailability struct {
	CaptionCount int                                   `json:"captionCount" example:"120"`
	HasOriginal  bool                                  `json:"hasOriginal" example:"true"`
	Translations []VideoCaptionTranslationAvailability `json:"translations"`
} // @name entity.VideoCaptionAvailability

type Video struct {
	ID                  int64                    `json:"id" example:"1"`
	Title               string                   `json:"title" example:"Learn English with Music"`
	YouTubeID           string                   `json:"youtubeId" example:"dQw4w9WgXcQ"`
	VideoURL            string                   `json:"videoUrl" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
	ThumbnailURL        string                   `json:"thumbnailUrl,omitempty"`
	DurationSeconds     int                      `json:"durationSeconds" example:"300"`
	Status              string                   `json:"status" example:"published"`
	Language            VideoLanguage            `json:"language"`
	Level               VideoLevel               `json:"level"`
	Channel             VideoChannel             `json:"channel"`
	Topics              []VideoTopic             `json:"topics"`
	CaptionAvailability VideoCaptionAvailability `json:"captionAvailability"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           time.Time                `json:"updatedAt"`
} // @name entity.Video

type VideoInput struct {
	Title           string
	YouTubeID       string
	ThumbnailURL    string
	DurationSeconds int
	Status          string
	LanguageID      int
	LevelID         int
	ChannelID       int
	TopicIDs        []int
}

type VideoList struct {
	Items []Video `json:"items"`
	Total int     `json:"total" example:"100"`
} // @name entity.VideoList

type UserVideo struct {
	ID                  int64
	Title               string
	YouTubeID           string
	ThumbnailURL        string
	DurationSeconds     int
	LevelCode           string
	LastPositionSeconds int
	LastWatchedAt       time.Time
	SavedAt             time.Time
}

type UserVideoList struct {
	Items []UserVideo
	Total int
}

type YouTubeVideoPreview struct {
	YouTubeID            string                 `json:"youtubeId" example:"dQw4w9WgXcQ"`
	Title                string                 `json:"title" example:"Video title"`
	ThumbnailURL         string                 `json:"thumbnailUrl,omitempty"`
	DurationSeconds      int                    `json:"durationSeconds" example:"300"`
	ChannelYouTubeID     string                 `json:"channelYoutubeId,omitempty"`
	ChannelName          string                 `json:"channelName,omitempty"`
	ManualSubtitleTracks []YouTubeSubtitleTrack `json:"manualSubtitleTracks"`
}
