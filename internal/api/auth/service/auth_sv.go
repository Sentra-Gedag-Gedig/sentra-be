package authService

import (
	"ProjectGolang/internal/api/auth"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	jwtPkg "ProjectGolang/pkg/jwt"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

func (s *authDomainImpl) Login(c context.Context, req auth.LoginUserRequest) (auth.LoginUserResponse, error) {
	requestID := contextPkg.GetRequestID(c)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return auth.LoginUserResponse{}, err
	}

	var user entity.User
	switch {
	case req.Email != "":
		user, err = repo.Users.GetByEmail(c, req.Email)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Warn("Failed to get user by email")

				err = auth.ErrInvalidEmailOrPassword
			} else {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to get user by email")

				return auth.LoginUserResponse{}, err
			}
		}
	case req.PhoneNumber != "":
		user, err = repo.Users.GetByPhoneNumber(c, req.PhoneNumber)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Warn("Failed to get user by phone number")
				err = auth.ErrInvalidEmailOrPassword
			} else {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to get user by phone number")

				return auth.LoginUserResponse{}, err
			}
		}
	default:
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Email or PhoneNumber is required",
		}).Warn("Invalid login request")
		return auth.LoginUserResponse{}, auth.ErrInvalidEmailOrPassword
	}

	if err := s.bcryptUtils.ComparePassword(user.Password, req.Password); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Password comparison failed")
		return auth.LoginUserResponse{}, auth.ErrInvalidEmailOrPassword
	}

	userData := MakeUserData(user)

	token, expired, err := jwtPkg.Sign(userData, time.Hour*1)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to sign token")
		return auth.LoginUserResponse{}, err
	}

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
	}).Info("Token created")

	res := auth.LoginUserResponse{
		AccessToken:      token,
		ExpiresInMinutes: time.Until(time.Unix(expired, 0)).Minutes(),
	}

	return res, nil
}

func (s *authDomainImpl) LoginGoogle() (*url.URL, error) {
	gConfig := s.googleProvider.GetConfig()
	URL, err := url.Parse(gConfig.Endpoint.AuthURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v", err)
		return nil, err
	}

	parameters := url.Values{}
	parameters.Add("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	parameters.Add("scope", strings.Join(gConfig.Scopes, " "))
	parameters.Add("redirect_uri", gConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", os.Getenv("GOOGLE_STATE"))
	URL.RawQuery = parameters.Encode()

	return URL, nil
}

func (s *authDomainImpl) UserLoginGoogle(c context.Context, req auth.LoginUserGoogle) (auth.LoginUserResponse, error) {
	requestID := contextPkg.GetRequestID(c)
	repo, err := s.repo.NewClient(false)
	if err != nil {

		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")

		return auth.LoginUserResponse{}, err
	}

	user, err := repo.Users.GetByEmail(c, req.Email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Failed to get user by email")

			err = auth.ErrInvalidEmailOrPassword
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to get user by email")

			return auth.LoginUserResponse{}, err
		}
	}

	userData := MakeUserData(user)

	token, expired, err := jwtPkg.Sign(userData, time.Hour*1)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to sign token")
		return auth.LoginUserResponse{}, err
	}

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
	}).Info("Token created")

	res := auth.LoginUserResponse{
		AccessToken:      token,
		ExpiresInMinutes: time.Until(time.Unix(expired, 0)).Minutes(),
	}

	return res, nil
}

func (s *authDomainImpl) PhoneNumberVerification(c context.Context, phoneNumber string) error {
	requestID := contextPkg.GetRequestID(c)

	verificationCode := fmt.Sprintf("%05d", 10000+rand.Intn(90000))
	if err := s.redisServer.SetOTP(c, phoneNumber, verificationCode, 1*time.Minute); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to set OTP in Redis")
		return err
	}

	if err := s.whatsappSender.SendMessage(c, phoneNumber, verificationCode); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to send WhatsApp message")
		return err
	}

	return nil
}

func (s *authDomainImpl) VerifyOTPandUpdatePIN(c context.Context, req auth.OTPPINRequest) error {
	requestID := contextPkg.GetRequestID(c)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")

	}

	storedOTP, err := s.redisServer.GetOTP(c, req.PhoneNumber)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get OTP from Redis")
		return auth.ErrorTokenExpired
	}

	if storedOTP != req.Code {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Invalid OTP",
		}).Warn("Invalid OTP")
		return auth.ErrorInvalidToken
	}

	hashedPIN, err := s.bcryptUtils.HashPassword(req.PIN)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to hash PIN")
		return err
	}

	if err := repo.Users.UpdateUserPIN(c, req.PhoneNumber, hashedPIN); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Failed to update user PIN")
			err = auth.ErrInvalidPhoneNumber
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to update user PIN")

			return err
		}
	}

	return nil
}

func (s *authDomainImpl) SendEmailOTP(c context.Context, email string) error {
	requestID := contextPkg.GetRequestID(c)

	// Generate a 5-digit OTP code
	verificationCode := fmt.Sprintf("%05d", 10000+rand.Intn(90000))

	// Store the OTP in Redis with a 5-minute expiration
	if err := s.redisServer.SetOTP(c, email, verificationCode, 5*time.Minute); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to set email OTP in Redis")
		return err
	}

	// Send the OTP via email
	if err := s.smtpMailer.CreateSmtp(email, verificationCode); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to send email OTP")
		return err
	}

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"email":      email,
	}).Info("Email OTP sent successfully")

	return nil
}

func (s *authDomainImpl) VerifyEmailOTP(c context.Context, userID string, email string, code string) error {
	requestID := contextPkg.GetRequestID(c)

	// Retrieve the stored OTP from Redis
	storedOTP, err := s.redisServer.GetOTP(c, email)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get OTP from Redis")
		return auth.ErrorTokenExpired
	}

	// Verify the OTP
	if storedOTP != code {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Invalid OTP",
		}).Warn("Invalid email OTP")
		return auth.ErrorInvalidToken
	}

	// Get a database client
	repo, err := s.repo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	// Check if email is already in use by another user
	existingUser, err := repo.Users.GetByEmail(c, email)
	if err == nil && existingUser.ID != userID {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"email":      email,
		}).Warn("Email already in use by another user")
		return auth.ErrEmailAlreadyInUse
	}

	// Get current user
	user, err := repo.Users.GetByID(c, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get user by ID")
		return err
	}

	// Update user email
	user.Email = email
	user.UpdatedAt = time.Now()

	// Save user with updated email
	if err := repo.Users.UpdateUser(c, user); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update user email")
		return err
	}

	// Commit the transaction
	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return err
	}

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    userID,
		"email":      email,
	}).Info("User email updated successfully")

	return nil
}

func (s *authDomainImpl) VerifyPhoneOTP(c context.Context, userID string, phoneNumber string, code string) error {
	requestID := contextPkg.GetRequestID(c)

	storedOTP, err := s.redisServer.GetOTP(c, phoneNumber)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get OTP from Redis")
		return auth.ErrorTokenExpired
	}

	if storedOTP != code {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Invalid OTP",
		}).Warn("Invalid phone OTP")
		return auth.ErrorInvalidToken
	}

	repo, err := s.repo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	existingUser, err := repo.Users.GetByPhoneNumber(c, phoneNumber)
	if err == nil && existingUser.ID != userID {
		s.log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"phoneNumber": phoneNumber,
		}).Warn("Phone number already in use by another user")
		return auth.ErrPhoneNumberAlreadyExists
	} else if !errors.Is(err, auth.ErrUserNotFound) && err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to check phone number usage")
		return err
	}

	user, err := repo.Users.GetByID(c, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get user by ID")
		return err
	}

	user.PhoneNumber = phoneNumber
	user.UpdatedAt = time.Now()

	if err := repo.Users.UpdateUser(c, user); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update user phone number")
		return err
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return err
	}

	s.log.WithFields(logrus.Fields{
		"request_id":  requestID,
		"user_id":     userID,
		"phoneNumber": phoneNumber,
	}).Info("User phone number updated successfully")

	return nil
}
