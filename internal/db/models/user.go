package models

type UserRole string

const (
	UserRoleGuest  UserRole = "guest"
	UserRoleMember UserRole = "member"
	UserRoleSeeder UserRole = "seeder"
)

type User struct {
	ID            int           `json:"id" pg:",pk,default:gen_random_uuid()"`
	TelegramID    int           `json:"telegram_id" pg:",notnull,unique"`
	DiscordID     int           `json:"discord_id"`
	Role          UserRole      `json:"role" pg:"type:UserRole,notnull,default:'guest'"`
	Proposals     []Proposal    `json:"proposals" pg:"rel:has-many"`
	TempProposal  Proposal      `json:"temp_proposal"`
	TelegramState TelegramState `json:"telegram_state"`
}
