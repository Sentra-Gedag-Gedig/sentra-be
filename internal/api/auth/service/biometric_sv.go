package authService

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	jwtPkg "ProjectGolang/pkg/jwt"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"time"
)

func (s *biometricDomainImpl) LoginTouchID(ctx context.Context, req auth.TouchIDLoginRequest) (auth.LoginUserResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return auth.LoginUserResponse{}, err
	}

	user, err := repo.Users.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			logrus.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("User not found")

			return auth.LoginUserResponse{}, auth.ErrInvalidEmailOrPassword
		}
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get user by ID")
		return auth.LoginUserResponse{}, err
	}

	if err := s.bcryptUtils.ComparePassword(user.HashTouchID, req.PlainText); err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to compare password")
		return auth.LoginUserResponse{}, auth.ErrInvalidEmailOrPassword
	}

	userData := MakeUserData(user)

	token, expired, err := jwtPkg.Sign(userData, time.Hour*1)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to sign token")
		return auth.LoginUserResponse{}, err
	}

	res := auth.LoginUserResponse{
		AccessToken:      token,
		ExpiresInMinutes: time.Until(time.Unix(expired, 0)).Minutes(),
	}

	return res, nil
}

func (s *biometricDomainImpl) EnableTouchID(ctx context.Context, userID string) (string, error) {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return "", err
	}

	ULID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate ULID")
		return "", err
	}

	hashedUlid, err := s.bcryptUtils.HashPassword(ULID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to hash password")
		return "", err
	}

	if err := repo.Users.EnableTouchID(ctx, userID, hashedUlid); err != nil {
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to enable Touch ID")
		return "", err
	}

	return ULID, nil
}
