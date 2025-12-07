package domain

import "time"

type Request struct {
	ID        string
	UserID    string
	ChannelID string
	ReqType   string
	Status    string
	CreatedAt time.Time
}
