package authService

import (
	"ProjectGolang/internal/api/auth"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"mime/multipart"
	"strings"
	"time"
)

func (s *userDomainImpl) RegisterUser(ctx context.Context, req auth.CreateUserRequest) error {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}

	hashedPassword, err := s.bcryptUtils.HashPassword(req.Password)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to hash password")
		return err
	}

	ULID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate ULID")
		return err
	}

	user := entity.User{
		ID:          ULID,
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
	}

	if err := repo.Users.CreateUser(ctx, user); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create user")
		return err
	}

	return nil
}

func (s *userDomainImpl) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return entity.User{}, err
	}

	user, err := repo.Users.GetByEmail(ctx, email)
	if err != nil {

		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("User not found")

			return entity.User{}, auth.ErrUserNotFound
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to get user by email")

			return entity.User{}, err
		}
	}

	return user, nil
}

func (s *userDomainImpl) UpdateUser(ctx context.Context, user entity.UserLoginData, req auth.UpdateUserRequest) error {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}

	userData, err := repo.Users.GetByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("User not found")

			return auth.ErrUserNotFound
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to get user by ID")

			return err
		}
	}

	newUser, err := GetUserDifferenceData(userData, req)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get user difference data")
		return err
	}

	if err := repo.Users.UpdateUser(ctx, newUser); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update user")
		return err
	}

	return nil
}

func (s *userDomainImpl) UpdateUserVerifiedStatusAndPIN(ctx context.Context, user auth.VerifyUserUsingOTP) error {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	storedOTP, err := s.redisServer.GetOTP(ctx, user.PhoneNumber)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get OTP from Redis")
		return auth.ErrorTokenExpired
	}

	if storedOTP != user.Code {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "Invalid OTP",
		}).Error("Invalid OTP provided")
		return auth.ErrInvalidOTP
	}

	userData, err := repo.Users.GetByPhoneNumber(ctx, user.PhoneNumber)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
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

	hashedPIN, err := s.bcryptUtils.HashPassword(user.PIN)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to hash PIN")
		return err
	}

	if err := repo.Users.UpdateUserPIN(ctx, user.PhoneNumber, hashedPIN); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update user PIN")
		return err
	}

	userData.IsVerified = true

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return err
	}

	return nil
}

func (s *userDomainImpl) DeleteUser(ctx context.Context, id string) error {
	requestID := contextPkg.GetRequestID(ctx)
	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}

	if err := repo.Users.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("User not found")

			return auth.ErrUserNotFound
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to delete user")

			return err
		}
	}

	return nil
}

func (s *userDomainImpl) UpdateProfilePhoto(ctx context.Context, userID string, photoFile *multipart.FileHeader) (*auth.ProfilePhotoResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	if photoFile == nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
		}).Warn("No file provided")
		return nil, auth.ErrInvalidFileType
	}

	if photoFile.Size > 5*1024*1024 {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"file_size":  photoFile.Size,
		}).Warn("File too large")
		return nil, auth.ErrFileTooLarge
	}

	contentType := photoFile.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"user_id":      userID,
			"content_type": contentType,
		}).Warn("Invalid file type")
		return nil, auth.ErrInvalidFileType
	}

	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	userData, err := repo.Users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    userID,
				"error":      err.Error(),
			}).Warn("User not found")
			return nil, auth.ErrUserNotFound
		}

		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get user by ID")
		return nil, err
	}

	uploadedFileURL, err := s.s3Client.UploadFile(photoFile)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to upload file to S3")
		return nil, auth.ErrFailedToUploadFile
	}

	if userData.ProfilePhotoURL != "" {
		oldPhotoURL := userData.ProfilePhotoURL
		parts := strings.Split(oldPhotoURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			go func() {
				if err := s.s3Client.DeleteFile(fileName); err != nil {
					s.log.WithFields(logrus.Fields{
						"request_id": requestID,
						"user_id":    userID,
						"file_name":  fileName,
						"error":      err.Error(),
					}).Warn("Failed to delete old profile photo")
				}
			}()
		}
	}

	if err := repo.Users.UpdateProfilePhoto(ctx, userID, uploadedFileURL); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to update profile photo URL in database")

		parts := strings.Split(uploadedFileURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			if err := s.s3Client.DeleteFile(fileName); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"user_id":    userID,
					"file_name":  fileName,
					"error":      err.Error(),
				}).Warn("Failed to delete uploaded file after database update failure")
			}
		}

		return nil, err
	}

	return &auth.ProfilePhotoResponse{
		ID:              userID,
		ProfilePhotoURL: uploadedFileURL,
	}, nil
}

func (s *userDomainImpl) UpdateFacePhoto(ctx context.Context, userID string, facePhotoFile *multipart.FileHeader) error {
	requestID := contextPkg.GetRequestID(ctx)

	if facePhotoFile == nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
		}).Warn("No file provided")
		return auth.ErrInvalidFileType
	}

	if facePhotoFile.Size > 5*1024*1024 {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"file_size":  facePhotoFile.Size,
		}).Warn("File too large")
		return auth.ErrFileTooLarge
	}

	contentType := facePhotoFile.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"user_id":      userID,
			"content_type": contentType,
		}).Warn("Invalid file type")
		return auth.ErrInvalidFileType
	}

	repo, err := s.repo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}

	userData, err := repo.Users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    userID,
				"error":      err.Error(),
			}).Warn("User not found")
			return auth.ErrUserNotFound
		}

		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get user by ID")
		return err
	}

	uploadedFileURL, err := s.s3Client.UploadFile(facePhotoFile)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to upload file to S3")
		return auth.ErrFailedToUploadFile
	}

	if userData.FacePhotoURL != "" {
		oldPhotoURL := userData.FacePhotoURL
		parts := strings.Split(oldPhotoURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			go func() {
				if err := s.s3Client.DeleteFile(fileName); err != nil {
					s.log.WithFields(logrus.Fields{
						"request_id": requestID,
						"user_id":    userID,
						"file_name":  fileName,
						"error":      err.Error(),
					}).Warn("Failed to delete old face photo")
				}
			}()
		}
	}

	if err := repo.Users.UpdateFacePhoto(ctx, userID, uploadedFileURL); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to update face photo URL in database")
		parts := strings.Split(uploadedFileURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			if err := s.s3Client.DeleteFile(fileName); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"user_id":    userID,
					"file_name":  fileName,
					"error":      err.Error(),
				}).Warn("Failed to delete uploaded file after database update failure")
			}
		}

		return err
	}
	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    userID,
	}).Info("Successfully updated face photo URL in database")
	return nil
}
