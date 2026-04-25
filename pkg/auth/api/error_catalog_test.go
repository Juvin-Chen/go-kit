package api

import (
	"errors"
	"testing"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

func TestResolveError(t *testing.T) {
	testCases := []struct {
		name              string
		inputError        error
		expectedCode      ErrorCode
		expectedRetryable bool
		expectedMessage   string
	}{
		{
			name:              "nil error",
			inputError:        nil,
			expectedCode:      ErrorCodeOK,
			expectedRetryable: false,
			expectedMessage:   "ok",
		},
		{
			name:              "invalid argument",
			inputError:        domain.ErrInvalidUserID,
			expectedCode:      ErrorCodeInvalidArgument,
			expectedRetryable: false,
			expectedMessage:   domain.ErrInvalidUserID.Error(),
		},
		{
			name:              "invalid token",
			inputError:        domain.ErrInvalidRefreshToken,
			expectedCode:      ErrorCodeInvalidToken,
			expectedRetryable: false,
			expectedMessage:   domain.ErrInvalidRefreshToken.Error(),
		},
		{
			name:              "session not found",
			inputError:        domain.ErrRefreshSessionNotFound,
			expectedCode:      ErrorCodeSessionNotFound,
			expectedRetryable: false,
			expectedMessage:   domain.ErrRefreshSessionNotFound.Error(),
		},
		{
			name:              "session conflict",
			inputError:        domain.ErrRefreshSessionConflict,
			expectedCode:      ErrorCodeSessionConflict,
			expectedRetryable: true,
			expectedMessage:   domain.ErrRefreshSessionConflict.Error(),
		},
		{
			name:              "session revoked",
			inputError:        domain.ErrRefreshSessionRevoked,
			expectedCode:      ErrorCodeSessionRevoked,
			expectedRetryable: false,
			expectedMessage:   domain.ErrRefreshSessionRevoked.Error(),
		},
		{
			name:              "session expired",
			inputError:        domain.ErrRefreshSessionExpired,
			expectedCode:      ErrorCodeSessionExpired,
			expectedRetryable: false,
			expectedMessage:   domain.ErrRefreshSessionExpired.Error(),
		},
		{
			name:              "unknown error",
			inputError:        errors.New("unknown"),
			expectedCode:      ErrorCodeInternal,
			expectedRetryable: false,
			expectedMessage:   app.ErrAuthInternal.Error(),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			result := ResolveError(testCase.inputError)
			if result.Code != testCase.expectedCode {
				t.Fatalf("expected code %s, got %s", testCase.expectedCode, result.Code)
			}
			if result.Retryable != testCase.expectedRetryable {
				t.Fatalf("expected retryable %v, got %v", testCase.expectedRetryable, result.Retryable)
			}
			if result.Message != testCase.expectedMessage {
				t.Fatalf("expected message %s, got %s", testCase.expectedMessage, result.Message)
			}
		})
	}
}
