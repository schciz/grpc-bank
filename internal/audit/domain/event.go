package domain

import "time"

type AuditEvent struct {
	ID        int64
	UserID    string
	Action    string
	IPAddress string
	Payload   string
	CreatedAt time.Time
}
