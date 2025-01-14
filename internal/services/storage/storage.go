package storage

import (
	"errors"
	"fmt"
)

var (
	ErrURLOrCodeExists = errors.New("url or code exists")
)

type ExistsURLError struct {
	OriginalURL string
	ShortCode   string
	Err         error
}

func (e *ExistsURLError) Error() string {
	return fmt.Sprintf("[%s:%s] %v", e.OriginalURL, e.ShortCode, e.Err)
}
