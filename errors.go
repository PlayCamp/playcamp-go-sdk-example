package main

import (
	"errors"
	"net/http"

	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleSDKError maps an SDK error to the appropriate HTTP status code and writes the error response.
func handleSDKError(w http.ResponseWriter, err error) {
	var (
		badReqErr     *playcamp.BadRequestError
		authErr       *playcamp.AuthError
		forbiddenErr  *playcamp.ForbiddenError
		notFoundErr   *playcamp.NotFoundError
		conflictErr   *playcamp.ConflictError
		validationErr *playcamp.ValidationError
		rateLimitErr  *playcamp.RateLimitError
		networkErr    *playcamp.NetworkError
		inputErr      *playcamp.InputValidationError
		apiErr        *playcamp.APIError
	)

	switch {
	case errors.As(err, &badReqErr):
		writeError(w, http.StatusBadRequest, badReqErr.Message)
	case errors.As(err, &authErr):
		writeError(w, http.StatusUnauthorized, authErr.Message)
	case errors.As(err, &forbiddenErr):
		writeError(w, http.StatusForbidden, forbiddenErr.Message)
	case errors.As(err, &notFoundErr):
		writeError(w, http.StatusNotFound, notFoundErr.Message)
	case errors.As(err, &conflictErr):
		writeError(w, http.StatusConflict, conflictErr.Message)
	case errors.As(err, &validationErr):
		writeError(w, http.StatusUnprocessableEntity, validationErr.Message)
	case errors.As(err, &rateLimitErr):
		writeError(w, http.StatusTooManyRequests, rateLimitErr.Message)
	case errors.As(err, &networkErr):
		writeError(w, http.StatusBadGateway, networkErr.Message)
	case errors.As(err, &inputErr):
		writeError(w, http.StatusBadRequest, inputErr.Error())
	case errors.As(err, &apiErr):
		writeError(w, apiErr.StatusCode, apiErr.Message)
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
