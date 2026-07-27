package request

import "time"

// TopicRequest -.
type TopicRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string `json:"slug" validate:"required,max=255"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
} // @name v1.TopicRequest

// LevelRequest -.
type LevelRequest struct {
	Code      string `json:"code" validate:"required,max=50"`
	Name      string `json:"name" validate:"required,max=255"`
	SortOrder int    `json:"sort_order"`
} // @name v1.LevelRequest

// ChannelRequest -.
type ChannelRequest struct {
	ChannelID    string `json:"channel_id" validate:"required,max=255"`
	ChannelName  string `json:"channel_name" validate:"required,max=255"`
	ThumbnailURL string `json:"thumbnail_url" validate:"omitempty,url"`
} // @name v1.ChannelRequest

// VideoRequest -.
type VideoRequest struct {
	VideoID      string     `json:"video_id" validate:"required,max=255"`
	ChannelID    int64      `json:"channel_id" validate:"required,gt=0"`
	Title        string     `json:"title" validate:"required,max=255"`
	Description  string     `json:"description"`
	ThumbnailURL string     `json:"thumbnail_url" validate:"omitempty,url"`
	ViewCount    int64      `json:"view_count" validate:"gte=0"`
	Duration     int        `json:"duration" validate:"gte=0"`
	SortOrder    int        `json:"sort_order"`
	IsActive     *bool      `json:"is_active"`
	PublishedAt  *time.Time `json:"published_at"`
	TopicIDs     []int64    `json:"topic_ids" validate:"unique,dive,gt=0"`
	LevelIDs     []int64    `json:"level_ids" validate:"unique,dive,gt=0"`
} // @name v1.VideoRequest
