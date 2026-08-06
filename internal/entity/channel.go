package entity

import "time"

type Channel struct {
	ID               int       `json:"id" example:"1"`
	ChannelYouTubeID string    `json:"channelYoutubeId" example:"UC_x5XG1OV2P6uZZ5FSM9Ttw"`
	ChannelName      string    `json:"channelName" example:"Google Developers"`
	AvatarURL        string    `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
	IsActive         bool      `json:"isActive" example:"true"`
	CreatedAt        time.Time `json:"createdAt" example:"2026-08-06T00:00:00Z"`
} // @name entity.Channel
