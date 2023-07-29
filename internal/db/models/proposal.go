package models

import "time"

type (
	ProposalStatus string
	NomineeRole    string
)

func (p ProposalStatus) String() string {
	return string(p)
}

func (r NomineeRole) String() string {
	return string(r)
}

const (
	ProposalStatusCreated  ProposalStatus = "created"
	ProposalStatusApproved ProposalStatus = "approved"
	ProposalStatusRejected ProposalStatus = "rejected"

	NomineeRoleMember NomineeRole = "member"
	NomineeRoleSeeder NomineeRole = "seeder"
)

type Proposal struct {
	ID                      int            `json:"id" pg:",pk,default:gen_random_uuid()"`
	NominatorID             int            `json:"nominator_id" pg:",notnull"`
	NomineeTelegramNickname string         `json:"nominee_telegram_nickname" pg:",notnull"`
	NomineeRole             NomineeRole    `json:"nominee_role" pg:",notnull"`
	Comment                 string         `json:"comment" pg:",notnull"`
	Status                  ProposalStatus `json:"status" pg:"type:ProposalStatus,notnull,default:'created'"`
	CreatedAt               time.Time      `json:"created_at" pg:"default:now()"`
	FinishedAt              time.Time      `json:"finished_at"`
	Votes                   []Vote         `json:"votes" pg:"rel:has-many"`
}
