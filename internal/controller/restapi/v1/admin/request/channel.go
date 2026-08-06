package request

type SaveChannel struct {
	ChannelYouTubeID string `json:"channelYoutubeId" validate:"required,max=100" example:"UC_x5XG1OV2P6uZZ5FSM9Ttw"`
	ChannelName      string `json:"channelName" validate:"required,max=150" example:"Google Developers"`
	AvatarURL        string `json:"avatarUrl" example:"https://example.com/avatar.jpg"`
	IsActive         *bool  `json:"isActive" validate:"required" example:"true"`
}
