package url

import (
	"context"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/utils/random"
)

type Service struct {
	Storage storage.URLStorage
}

func (s *Service) Create(ctx context.Context, sourceURL string) (string, error) {
	mURL, err := s.Storage.GetUrlByUrl(ctx, sourceURL)
	if err != nil {
		return "", err
	}
	if mURL.ID > 0 {
		return mURL.Code, nil
	}

	var code string
	isUnique := false

	for !isUnique {
		code = random.GenerateRandomString(10)

		mURL, err = s.Storage.GetUrlById(ctx, code)
		if err != nil {
			return "", err
		}
		if mURL.ID == 0 {
			isUnique = true
		}
	}

	_, err = s.Storage.SaveUrl(ctx, code, sourceURL)

	if err != nil {
		return "", err
	}

	return code, nil
}
