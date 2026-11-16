package domain

import "time"

type StatementTransaction struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        string
	Status        string
	CreatedAt     time.Time
}

type StatementReport struct {
	AccountID   string
	StartDate   string
	EndDate     string
	Format      string
	DownloadURL string
	Storage     string
	ExpiresIn   string
}
