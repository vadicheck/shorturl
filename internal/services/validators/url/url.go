package url

import "net/url"

func IsValid(rawURL string) (bool, error) {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false, err
	}
	return true, nil
}
