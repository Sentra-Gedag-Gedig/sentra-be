package auth

import "time"

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=13"`
	Password    string `json:"password" validate:"required,min=8,max=32"`
}

type VerifyPhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=13"`
}

type OTPPINRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=13"`
	Code        string `json:"code" validate:"required,min=5,max=5"`
	PIN         string `json:"personal_identification_number" validate:"required,min=6,max=6"`
}

type VerifyUserUsingOTP struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=13"`
	Code        string `json:"code" validate:"required,min=5,max=5"`
	PIN         string `json:"personal_identification_number" validate:"required,min=6,max=6"`
}

type UpdateProfilePhotoRequest struct {
	ID string `json:"id"`
}

type ProfilePhotoResponse struct {
	ID              string `json:"id"`
	ProfilePhotoURL string `json:"profile_photo_url"`
}

type UserResponse struct {
	ID                        string    `json:"id"`
	Email                     string    `json:"email,omitempty"`
	Name                      string    `json:"name"`
	NationalIdentityNumber    string    `json:"national_identity_number,omitempty"`
	BirthPlace                string    `json:"birth_place,omitempty"`
	BirthDate                 time.Time `json:"birth_date,omitempty"`
	Gender                    string    `json:"gender,omitempty"`
	Address                   string    `json:"address,omitempty"`
	NeighborhoodCommunityUnit string    `json:"neighborhood_community_unit,omitempty"`
	Village                   string    `json:"village,omitempty"`
	District                  string    `json:"district,omitempty"`
	Religion                  string    `json:"religion,omitempty"`
	MaritalStatus             string    `json:"marital_status,omitempty"`
	Profession                string    `json:"profession,omitempty"`
	Citizenship               string    `json:"citizenship,omitempty"`
	CardValidUntil            time.Time `json:"card_valid_until,omitempty"`
	PhoneNumber               string    `json:"phone_number,omitempty"`
	ProfilePhotoURL           string    `json:"profile_photo_url,omitempty"`
	IsVerified                bool      `json:"is_verified"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type LoginUserRequest struct {
	Email       string `json:"email" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10,max=13"`
	Password    string `json:"password" validate:"required"`
}

type TouchIDLoginRequest struct {
	ID        string `json:"id"`
	PlainText string `json:"plain_text"`
}

type ResetPassword struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
	Password    string `json:"password" validate:"required,min=8,max=32"`
}

type LoginUserGoogle struct {
	Email string `json:"email"`
}

type UserGoogle struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

type LoginUserResponse struct {
	AccessToken      string  `json:"accessToken"`
	ExpiresInMinutes float64 `json:"expiresInHour"`
}

type UpdateUserRequest struct {
	Name                      string `json:"name"`
	NationalIdentityNumber    string `json:"national_identity_number"`
	BirthPlace                string `json:"birth_place"`
	BirthDate                 string `json:"birth_date"`
	Gender                    string `json:"gender"`
	Address                   string `json:"address"`
	NeighborhoodCommunityUnit string `json:"neighborhood_community_unit"`
	Village                   string `json:"village"`
	District                  string `json:"district"`
	Religion                  string `json:"religion"`
	MaritalStatus             string `json:"marital_status"`
	Profession                string `json:"profession"`
	Citizenship               string `json:"citizenship"`
	CardValidUntil            string `json:"card_valid_until"`
}

type UpdateUserPINRequest struct {
	PhoneNumber string `json:"phone_number"`
	PIN         string `json:"personal_identification_number"`
}

type SendEmailOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyEmailOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=5,max=5"`
}

type VerifyPhoneOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=13"`
	Code        string `json:"code" validate:"required,min=5,max=5"`
}
