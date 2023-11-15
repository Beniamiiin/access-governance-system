package configs

type Discord struct {
	Token        string `env:"DISCORD_AUTHORIZATION_BOT_TOKEN"`
	ChannelID    string `env:"DISCORD_SERVER_ID"`
	GuestRoleID  string `env:"DISCORD_GUEST_ROLE_ID"`
	MemberRoleID string `env:"DISCORD_MEMBER_ROLE_ID"`
}
