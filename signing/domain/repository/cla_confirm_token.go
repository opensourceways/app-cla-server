package repository

import "time"

// CLAConfirmTokenPayload is the server-side data bound to a one-time
// cla confirm token. It is stored in redis and never exposed to clients.
type CLAConfirmTokenPayload struct {
	LinkId    string
	Email     string
	NewCLAId  string
	IssuedAt  time.Time
}

// CLAConfirmToken manages one-time tokens for confirming individual CLA
// updates. A token is bound to (linkId, email, newClaId), expires after
// ttl, and can be consumed only once.
type CLAConfirmToken interface {
	Add(linkId, email, newClaId string, ttl time.Duration) (token string, err error)
	Consume(token string) (CLAConfirmTokenPayload, error)
}
