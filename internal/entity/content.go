package entity

import "time"

// Topic -.
type Topic struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
} // @name entity.Topic

// Level -.
type Level struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
} // @name entity.Level

// Channel -.
type Channel struct {
	ID           int64     `json:"id"`
	ChannelID    string    `json:"channel_id"`
	ChannelName  string    `json:"channel_name"`
	ThumbnailURL string    `json:"thumbnail_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
} // @name entity.Channel

// Video -.
type Video struct {
	ID           int64      `json:"id"`
	VideoID      string     `json:"video_id"`
	ChannelID    int64      `json:"channel_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ThumbnailURL string     `json:"thumbnail_url"`
	ViewCount    int64      `json:"view_count"`
	Duration     int        `json:"duration"`
	SortOrder    int        `json:"sort_order"`
	IsActive     bool       `json:"is_active"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	TopicIDs     []int64    `json:"topic_ids"`
	LevelIDs     []int64    `json:"level_ids"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
} // @name entity.Video
