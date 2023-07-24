package models

import "time"

type (
	ProposalStatus string
	NomineeRole    string
)

const (
	ProposalStatusCreated  ProposalStatus = "created"
	ProposalStatusApproved ProposalStatus = "approved"
	ProposalStatusDeclined ProposalStatus = "declined"

	NomineeRoleMember NomineeRole = "member"
	NomineeRoleSeeder NomineeRole = "seeder"
)

type Proposal struct {
	ID                      int            `json:"id" pg:",pk,default:gen_random_uuid()"`
	NominatorID             int            `json:"nominator_id" pg:",notnull"`
	NomineeTelegramNickname string         `json:"nominee_telegram_nickname" pg:",notnull"`
	NomineeTelegramID       int            `json:"nominee_telegram_id" pg:",notnull"`
	NomineeRole             NomineeRole    `json:"nominee_role" pg:",notnull"`
	Description             string         `json:"description" pg:",notnull"`
	Status                  ProposalStatus `json:"status" pg:"type:ProposalStatus,notnull,default:'creating'"`
	CreatedAt               time.Time      `json:"created_at" pg:"default:now()"`
	Votes                   []Vote         `json:"votes" pg:"rel:has-many"`
}
