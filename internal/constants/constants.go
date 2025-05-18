// Package constants defines various constant values used throughout the application.
//
// It includes constants for HTTP header keys and other fixed values needed in the code.
package constants

// headerKey represents a custom type for HTTP header keys.
type headerKey string

// mdKey represents a custom type for HTTP header keys.
type mdKey string

// XUserID is the key used in HTTP headers to represent the user ID.
//
// This constant is used to access or set the "X-User-ID" header in HTTP requests.
const XUserID headerKey = "X-User-ID"

// MdUserID is the key used in gRPC metadata to represent the user ID.
//
// This constant is used to access or set the "x-user-id" header in gRPC requests.
const MdUserID mdKey = "x-user-id"
