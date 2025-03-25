// Package url provides a function for URL validation.
package url

import "net/url"

// IsValid checks whether a given string is a valid URL.
//
// The function uses `url.ParseRequestURI` to parse the string and verify its validity.
//
// Parameters:
//   - rawURL: the string containing the URL to be checked.
//
// Returns:
//   - `true` if the URL is valid;
//   - `false` and an error if the string is not a valid URL.
//
// Example usage:
//
//	isValid, err := IsValid("https://example.com")
//	if err != nil {
//	    fmt.Println("Invalid URL:", err)
//	}
//	fmt.Println("Valid URL:", isValid)
func IsValid(rawURL string) (bool, error) {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false, err
	}
	return true, nil
}
