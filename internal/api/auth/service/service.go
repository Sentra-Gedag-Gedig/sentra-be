package authService

import (
	"ProjectGolang/internal/api/auth"
	authRepository "ProjectGolang/internal/api/auth/repository"
	"ProjectGolang/internal/entity"
	"ProjectGolang/pkg/bcrypt"
	"ProjectGolang/pkg/google"
	"ProjectGolang/pkg/redis"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/smtp"
	"ProjectGolang/pkg/utils"
	"ProjectGolang/pkg/whatsapp"
	"context"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"net/url"
)

type AuthService interface {
	User() UserDomain
	Auth() AuthDomain
	Password() PasswordDomain
	Biometric() BiometricDomain
	GetRepository() authRepository.Repository
}

type UserDomain interface {
	RegisterUser(c context.Context, req auth.CreateUserRequest) error
	GetByEmail(c context.Context, email string) (entity.User, error)
	UpdateUser(c context.Context, user entity.UserLoginData, req auth.UpdateUserRequest) error
	UpdateUserVerifiedStatusAndPIN(ctx context.Context, user auth.VerifyUserUsingOTP) error
	DeleteUser(c context.Context, id string) error
	UpdateProfilePhoto(c context.Context, userID string, photoFile *multipart.FileHeader) (*auth.ProfilePhotoResponse, error)
	UpdateFacePhoto(ctx context.Context, userID string, facePhotoFile *multipart.FileHeader) error
}

type AuthDomain interface {
	Login(c context.Context, req auth.LoginUserRequest) (auth.LoginUserResponse, error)
	LoginGoogle() (*url.URL, error)
	UserLoginGoogle(c context.Context, req auth.LoginUserGoogle) (auth.LoginUserResponse, error)
	PhoneNumberVerification(c context.Context, phoneNumber string) error
	VerifyOTPandUpdatePIN(c context.Context, req auth.OTPPINRequest) error
	SendEmailOTP(c context.Context, email string) error
	VerifyEmailOTP(c context.Context, userID string, email string, code string) error
	VerifyPhoneOTP(c context.Context, userID string, phoneNumber string, code string) error
}

type PasswordDomain interface {
	UpdatePassword(c context.Context, req auth.ResetPassword) error
}

type BiometricDomain interface {
	EnableTouchID(c context.Context, userID string) (string, error)
	LoginTouchID(c context.Context, req auth.TouchIDLoginRequest) (auth.LoginUserResponse, error)
}

type authService struct {
	log            *logrus.Logger
	authRepository authRepository.Repository
	googleProvider google.ItfGoogle
	smtpMailer     smtp.ItfSmtp
	redisServer    redis.IRedis
	whatsappSender whatsapp.IWhatsappSender
	s3Client       s3.ItfS3
	smtpClient     smtp.ItfSmtp
	bcryptUtils    bcrypt.IBcrypt
	utils          utils.IUtils

	userDomain      UserDomain
	authDomain      AuthDomain
	passwordDomain  PasswordDomain
	biometricDomain BiometricDomain
}

func (a *authService) User() UserDomain {
	return a.userDomain
}

func (a *authService) Auth() AuthDomain {
	return a.authDomain
}

func (a *authService) Password() PasswordDomain {
	return a.passwordDomain
}

func (a *authService) Biometric() BiometricDomain {
	return a.biometricDomain
}

func (a *authService) GetRepository() authRepository.Repository {
	return a.authRepository
}

type userDomainImpl struct {
	log         *logrus.Logger
	repo        authRepository.Repository
	redisServer redis.IRedis
	s3Client    s3.ItfS3
	smtpCLient  smtp.ItfSmtp
	bcryptUtils bcrypt.IBcrypt
	utils       utils.IUtils
}

type authDomainImpl struct {
	log            *logrus.Logger
	repo           authRepository.Repository
	googleProvider google.ItfGoogle
	redisServer    redis.IRedis
	whatsappSender whatsapp.IWhatsappSender
	smtpMailer     smtp.ItfSmtp
	bcryptUtils    bcrypt.IBcrypt
}

type passwordDomainImpl struct {
	log         *logrus.Logger
	repo        authRepository.Repository
	smtpMailer  smtp.ItfSmtp
	redisServer redis.IRedis
	bcryptUtils bcrypt.IBcrypt
}

type biometricDomainImpl struct {
	log         *logrus.Logger
	repo        authRepository.Repository
	redisServer redis.IRedis
	bcryptUtils bcrypt.IBcrypt
	utils       utils.IUtils
}

func New(log *logrus.Logger,
	authRepo authRepository.Repository,
	googleProvider google.ItfGoogle,
	smtpMailer smtp.ItfSmtp,
	redisServer redis.IRedis,
	whatsappSender whatsapp.IWhatsappSender,
	s3Client s3.ItfS3,
	smtp smtp.ItfSmtp,
	bcryptUtils bcrypt.IBcrypt,
	utils utils.IUtils,
) AuthService {
	return &authService{
		log:            log,
		authRepository: authRepo,
		googleProvider: googleProvider,
		smtpMailer:     smtpMailer,
		redisServer:    redisServer,
		whatsappSender: whatsappSender,
		s3Client:       s3Client,
		smtpClient:     smtp,
		bcryptUtils:    bcryptUtils,
		utils:          utils,

		userDomain:      &userDomainImpl{log: log, repo: authRepo, redisServer: redisServer, s3Client: s3Client, smtpCLient: smtp, bcryptUtils: bcryptUtils, utils: utils},
		authDomain:      &authDomainImpl{log: log, repo: authRepo, googleProvider: googleProvider, redisServer: redisServer, whatsappSender: whatsappSender, smtpMailer: smtpMailer, bcryptUtils: bcryptUtils},
		passwordDomain:  &passwordDomainImpl{log: log, repo: authRepo, smtpMailer: smtpMailer, redisServer: redisServer, bcryptUtils: bcryptUtils},
		biometricDomain: &biometricDomainImpl{log: log, repo: authRepo, redisServer: redisServer, bcryptUtils: bcryptUtils, utils: utils},
	}
}
