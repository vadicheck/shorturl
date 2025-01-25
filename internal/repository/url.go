package repository

type BatchURL struct {
	CorrelationID string
	ShortCode     string
}

type BatchURLDto struct {
	CorrelationID string
	OriginalURL   string
	ShortCode     string
}
