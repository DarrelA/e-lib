package repository

import "github.com/DarrelA/e-lib/internal/apperrors"

type SessionRepository interface {
	NewSession(userID int64) (string, *apperrors.RestErr)
}
