// Package models defines the data structures used in the application, including the URL model.
package models

// URL represents a shortened URL entry in the database.
// It contains information about the original URL, its shortened code, and the associated user ID.
type URL struct {
	// ID is the unique identifier for the URL entry in the database.
	ID int64 `json:"id"`

	// Code is the shortened code associated with the URL.
	Code string `json:"code"`

	// URL is the original URL that has been shortened.
	URL string `json:"url"`

	// UserID is the ID of the user who created the shortened URL.
	UserID string `json:"user_id"`

	// IsDeleted indicates whether the URL has been deleted.
	// If true, the URL has been marked as deleted.
	IsDeleted bool `json:"is_deleted"`
}
