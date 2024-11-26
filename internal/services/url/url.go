package url

import (
	"context"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/utils/random"
)

type Service struct {
	Storage storage.UrlStorage
}

func (s *Service) Create(ctx context.Context, sourceUrl string) (string, error) {
	mUrl, err := s.Storage.GetUrlByUrl(ctx, sourceUrl)
	if err != nil {
		return "", err
	}
	if mUrl.ID > 0 {
		return mUrl.Code, nil
	}

	var code string
	isUnique := false

	for !isUnique {
		code = random.GenerateRandomString(10)

		mUrl, err = s.Storage.GetUrlById(ctx, code)
		if err != nil {
			return "", err
		}
		if mUrl.ID == 0 {
			isUnique = true
		}
	}

	_, err = s.Storage.SaveUrl(ctx, code, sourceUrl)

	if err != nil {
		return "", err
	}

	return code, nil
}
