package models

import "time"

type VoteType string

const (
	VoteTypeYes         VoteType = "yes"
	VoteTypeNow         VoteType = "no"
	VoteTypeAcknowledge VoteType = "acknowledge"
)

type Vote struct {
	ID         int       `json:"id" pg:",pk,default:gen_random_uuid()"`
	UserID     int       `json:"user_id" pg:",notnull"`
	ProposalID string    `json:"proposal_id" pg:",notnull"`
	Type       VoteType  `json:"type" pg:"type:VoteType,notnull"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at" pg:"default:now()"`
}
