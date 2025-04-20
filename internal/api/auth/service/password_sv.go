package authService

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func (s *passwordDomainImpl) UpdatePassword(c context.Context, req auth.ResetPassword) error {
	requestID := contextPkg.GetRequestID(c)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}

	storedOTP, err := s.redisServer.GetOTP(c, req.PhoneNumber)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get OTP from Redis")
		return err
	}

	if storedOTP != req.Code {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Invalid OTP",
		}).Error("Invalid OTP provided")
		return auth.ErrInvalidOTP
	}

	user, err := repo.Users.GetByPhoneNumber(c, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      "User not found",
			}).Warn("User not found")
			return auth.ErrUserNotFound
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to get user by phone number")

			return err
		}
	}

	if req.Password == user.Password {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "New password cannot be the same as the old password",
		}).Error("New password is the same as the old password")
		return auth.ErrPasswordSame
	}

	hashedPass, err := s.bcryptUtils.HashPassword(req.Password)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to hash password")
		return err
	}

	if err := repo.Users.UpdateUserPassword(c, req.PhoneNumber, hashedPass); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update user password")
		if errors.Is(err, auth.ErrUserNotFound) {
			return auth.ErrInvalidPhoneNumber
		}
		return err
	}

	return nil
}
