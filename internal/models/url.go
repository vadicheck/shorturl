package models

type URL struct {
	ID        int64
	Code      string
	URL       string
	UserID    string
	IsDeleted bool
}
