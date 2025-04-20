package entity

import "time"

type User struct {
	ID                           string    `db:"id"`
	Email                        string    `db:"email"`
	Name                         string    `db:"name"`
	NationalIdentityNumber       string    `db:"national_identity_number"`
	BirthPlace                   string    `db:"birth_place"`
	BirthDate                    time.Time `db:"birth_date"`
	Gender                       string    `db:"gender"`
	Address                      string    `db:"address"`
	NeighborhoodCommunityUnit    string    `db:"neighborhood_community_unit"`
	Village                      string    `db:"village"`
	District                     string    `db:"district"`
	Religion                     string    `db:"religion"`
	MaritalStatus                string    `db:"marital_status"`
	Profession                   string    `db:"profession"`
	Citizenship                  string    `db:"citizenship"`
	CardValidUntil               time.Time `db:"card_valid_until"`
	Password                     string    `db:"password"`
	PhoneNumber                  string    `db:"phone_number"`
	PersonalIdentificationNumber string    `db:"personal_identification_number"`
	EnableTouchID                bool      `db:"enable_touch_id"`
	HashTouchID                  string    `db:"hash_touch_id"`
	ProfilePhotoURL              string    `db:"profile_photo_url"`
	FacePhotoURL                 string    `db:"face_photo_url"`
	IsVerified                   bool      `db:"is_verified"`
	CreatedAt                    time.Time `db:"created_at"`
	UpdatedAt                    time.Time `db:"updated_at"`
}

type UserLoginData struct {
	ID       string
	Username string
	Email    string
}

type PositionStatus string

const (
	NoFaceDetected  PositionStatus = "NO_FACE_DETECTED"
	PerfectPosition PositionStatus = "PERFECT_POSITION"
	AdjustPosition  PositionStatus = "ADJUST_POSITION"
)

type FacePosition struct {
	Status       PositionStatus `json:"status"`
	Instructions []string       `json:"instructions,omitempty"`
	XDeviation   float64        `json:"x_deviation,omitempty"`
	YDeviation   float64        `json:"y_deviation,omitempty"`
	FaceRatio    float64        `json:"face_ratio,omitempty"`
}

type Frame struct {
	Data   []byte
	Width  int
	Height int
}
