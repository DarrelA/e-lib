package postgres

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

type UserRepository interface {
	GetUser(provider string, id string, email string) (int, *apperrors.RestErr)
	SaveUser(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr)
}
