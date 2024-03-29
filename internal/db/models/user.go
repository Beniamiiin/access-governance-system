package models

type UserRole string

const (
	UserRoleGuest  UserRole = "guest"
	UserRoleMember UserRole = "member"
	UserRoleSeeder UserRole = "seeder"
)

type TelegramState struct {
	LastCommand      string
	LastCommandState string
}

func (r UserRole) String() string {
	return string(r)
}

type User struct {
	ID                    int           `json:"id" pg:",pk,default:gen_random_uuid()"`
	Name                  string        `json:"name" pg:",notnull"`
	TelegramID            int64         `json:"telegram_id" pg:",notnull,unique"`
	TelegramNickname      string        `json:"telegram_nickname" pg:",notnull,unique"`
	DiscordID             int           `json:"discord_id"`
	Role                  UserRole      `json:"role" pg:"type:UserRole,notnull,default:'guest'"`
	Proposals             []Proposal    `json:"proposals" pg:"rel:has-many,fk:user_id"`
	BackersID             []int64       `json:"backers_id" pg:",array"`
	NominatorID           int           `json:"nominator_id"`
	MembersChatInviteLink string        `json:"members_chat_invite_link"`
	SeedersChatInviteLink string        `json:"seeders_chat_invite_link"`
	TempProposal          Proposal      `json:"temp_proposal"`
	TelegramState         TelegramState `json:"telegram_state"`
}
