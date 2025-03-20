package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

type SessionRepository interface {
	NewSession(userID int64) (string, *apperrors.RestErr)
	GetSessionData(sessionID string) (*entity.Session, *apperrors.RestErr)
}
