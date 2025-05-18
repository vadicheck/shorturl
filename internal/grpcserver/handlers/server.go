// Package handlers provides gRPC handlers for administrative operations,
// such as health checks, URL management, and statistics reporting.
package handlers

import (
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/internal/validator"
)

// ServerAdmin implements gRPC handlers for administrative operations.
//
// It provides methods for managing shortened URLs, retrieving user-specific data,
// performing internal statistics checks, and handling batch operations.
type ServerAdmin struct {
	storage    urlservice.URLStorage
	validator  *validator.Validator
	urlService *urlservice.Service
}

// NewServer creates a new instance of ServerAdmin.
//
// It initializes the ServerAdmin with the given storage, validator, and URL service.
// This server is responsible for handling admin-related gRPC methods.
//
// Parameters:
// - storage: Implementation of the URLStorage interface.
// - validator: Pointer to the request validator.
// - urlService: Pointer to the URL business logic service.
//
// Returns:
// - *ServerAdmin: A fully initialized ServerAdmin instance ready to register gRPC handlers.
func NewServer(
	storage urlservice.URLStorage,
	validator *validator.Validator,
	urlService *urlservice.Service,
) *ServerAdmin {
	return &ServerAdmin{
		storage:    storage,
		validator:  validator,
		urlService: urlService,
	}
}
