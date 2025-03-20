package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

type UserRepository interface {
	GetUserByID(userID int64) (*dto.UserDetail, *apperrors.RestErr)
	GetUserID(provider string, id string, email string) (int64, *apperrors.RestErr)
	SaveUser(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr)
}
