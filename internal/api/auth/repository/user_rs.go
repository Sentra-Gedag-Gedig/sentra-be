package authRepository

import (
	"ProjectGolang/internal/api/auth"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"time"
)

type UserDB struct {
	ID                           sql.NullString `db:"id"`
	Email                        sql.NullString `db:"email"`
	Name                         sql.NullString `db:"name"`
	NationalIdentityNumber       sql.NullString `db:"national_identity_number"`
	BirthPlace                   sql.NullString `db:"birth_place"`
	BirthDate                    sql.NullTime   `db:"birth_date"`
	Gender                       sql.NullString `db:"gender"`
	Address                      sql.NullString `db:"address"`
	NeighborhoodCommunityUnit    sql.NullString `db:"neighborhood_community_unit"`
	Village                      sql.NullString `db:"village"`
	District                     sql.NullString `db:"district"`
	Religion                     sql.NullString `db:"religion"`
	MaritalStatus                sql.NullString `db:"marital_status"`
	Profession                   sql.NullString `db:"profession"`
	Citizenship                  sql.NullString `db:"citizenship"`
	CardValidUntil               sql.NullTime   `db:"card_valid_until"`
	Password                     sql.NullString `db:"password"`
	PhoneNumber                  sql.NullString `db:"phone_number"`
	PersonalIdentificationNumber sql.NullString `db:"personal_identification_number"`
	EnableTouchID                bool           `db:"enable_touch_id"`
	HashTouchID                  sql.NullString `db:"hash_touch_id"`
	ProfilePhotoURL              sql.NullString `db:"profile_photo_url"`
	FacePhotoURL                 sql.NullString `db:"face_photo_url"`
	IsVerified                   bool           `db:"is_verified"`
	CreatedAt                    sql.NullTime   `db:"created_at"`
	UpdatedAt                    sql.NullTime   `db:"updated_at"`
}

func (r *userRepository) CreateUser(c context.Context, user entity.User) error {
	requestID := contextPkg.GetRequestID(c)
	argsKV := map[string]interface{}{
		"id":           user.ID,
		"phone_number": user.PhoneNumber,
		"name":         user.Name,
		"password":     user.Password,
		"created_at":   time.Now(),
	}

	query, args, err := sqlx.Named(queryCreateUser, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateUser")
		return err
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(c, query, args...)
	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				if pqErr.Constraint == "users_phone_number_key" {
					r.log.WithFields(logrus.Fields{
						"request_id": requestID,
						"error":      err.Error(),
					}).Warn("Phone number already exists")
					return auth.ErrPhoneNumberAlreadyExists
				}
			}
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating user")

		return err
	}

	return nil
}

func (r *userRepository) GetByID(c context.Context, id string) (entity.User, error) {
	requestID := contextPkg.GetRequestID(c)
	var user UserDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetById, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByID named query preparation err")

		return entity.User{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(c, query, args...).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetByID no rows found")
			return entity.User{}, auth.ErrUserNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByID execution err")
		return entity.User{}, err
	}

	userRes := r.makeUser(user)

	return userRes, nil
}

func (r *userRepository) GetByPhoneNumber(c context.Context, phoneNumber string) (entity.User, error) {
	requestID := contextPkg.GetRequestID(c)
	var user UserDB

	argsKV := map[string]interface{}{
		"phone_number": phoneNumber,
	}

	query, args, err := sqlx.Named(queryGetByPhoneNumber, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByPhoneNumber named query preparation err")
		return entity.User{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(c, query, args...).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetByPhoneNumber no rows found")
			return entity.User{}, auth.ErrUserNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByPhoneNumber execution err")
		return entity.User{}, err
	}

	userRes := r.makeUser(user)
	return userRes, nil
}

func (r *userRepository) UpdateUser(c context.Context, user entity.User) error {
	requestID := contextPkg.GetRequestID(c)
	argsKV := map[string]interface{}{
		"id":                          user.ID,
		"name":                        user.Name,
		"national_identity_number":    user.NationalIdentityNumber,
		"birth_place":                 user.BirthPlace,
		"birth_date":                  user.BirthDate,
		"gender":                      user.Gender,
		"address":                     user.Address,
		"neighborhood_community_unit": user.NeighborhoodCommunityUnit,
		"village":                     user.Village,
		"district":                    user.District,
		"religion":                    user.Religion,
		"marital_status":              user.MaritalStatus,
		"profession":                  user.Profession,
		"citizenship":                 user.Citizenship,
		"card_valid_until":            user.CardValidUntil,
		"is_verified":                 user.IsVerified,
		"updated_at":                  time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateUser, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateUser named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(c, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("UpdateUser no rows found")
			return auth.ErrUserNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateUser execution err")

		return err
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryDeleteUser, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteUser named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("DeleteUser no rows found")

			return auth.ErrUserNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteUser execution err")

		return err
	}

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var user entity.User

	argsKV := map[string]interface{}{
		"email": email,
	}

	query, args, err := sqlx.Named(queryGetByEmail, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByEmail named query preparation err")

		return entity.User{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetByEmail no rows found")

			return entity.User{}, auth.ErrUserWithEmailNotFound
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetByEmail execution err")

		return entity.User{}, err
	}

	return user, nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, phoneNum string, password string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"phone_number": phoneNum,
		"password":     password,
	}

	query, args, err := sqlx.Named(queryUpdateUserPassword, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateUserPassword named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("UpdateUserPassword no rows found")

			return auth.ErrUserWithEmailNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateUserPassword execution err")
		return err
	}

	return nil
}

func (r *userRepository) UpdateUserPIN(ctx context.Context, phoneNum string, pin string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"phone_number":                   phoneNum,
		"is_verified":                    true,
		"personal_identification_number": pin,
		"updated_at":                     time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateUserVerificationPIN, argsKV)
	if err != nil {
		r.log.WithFields(
			logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("UpdateUserPIN named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("UpdateUserPIN no rows found")
			return auth.ErrInvalidPhoneNumber
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateUserPIN execution err")
		return err
	}

	return nil
}

func (r *userRepository) EnableTouchID(ctx context.Context, id string, hash string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id":              id,
		"enable_touch_id": true,
		"hash_touch_id":   hash,
	}

	query, args, err := sqlx.Named(queryUpdateEnableTouchID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("EnableTouchID named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("EnableTouchID no rows found")

			return auth.ErrUserWithEmailNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("EnableTouchID execution err")

		return err
	}

	return nil
}

func (r *userRepository) UpdateProfilePhoto(ctx context.Context, id string, photoURL string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id":                id,
		"profile_photo_url": photoURL,
		"updated_at":        time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateProfilePhoto, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateProfilePhoto named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("UpdateProfilePhoto no rows found")
			return auth.ErrUserNotFound
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateProfilePhoto execution err")
		return err
	}

	return nil
}

func (r *userRepository) UpdateFacePhoto(ctx context.Context, id string, facePhotoURL string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id":             id,
		"face_photo_url": facePhotoURL,
		"updated_at":     time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateFacePhoto, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateFacePhoto named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	if _, err := r.q.ExecContext(ctx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("UpdateFacePhoto no rows found")
			return auth.ErrUserNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateFacePhoto execution err")
		return err
	}

	return nil
}

func (r *userRepository) makeUser(user UserDB) entity.User {
	var createdAt, updatedAt time.Time

	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time
	}

	if user.UpdatedAt.Valid {
		updatedAt = user.UpdatedAt.Time
	}

	var birthDate time.Time
	if user.BirthDate.Valid {
		birthDate = user.BirthDate.Time
	}

	var cardValidUntil time.Time
	if user.CardValidUntil.Valid {
		cardValidUntil = user.CardValidUntil.Time
	}

	userRes := entity.User{
		ID:                           user.ID.String,
		Email:                        user.Email.String,
		Name:                         user.Name.String,
		NationalIdentityNumber:       user.NationalIdentityNumber.String,
		BirthPlace:                   user.BirthPlace.String,
		BirthDate:                    birthDate,
		Gender:                       user.Gender.String,
		Address:                      user.Address.String,
		NeighborhoodCommunityUnit:    user.NeighborhoodCommunityUnit.String,
		Village:                      user.Village.String,
		District:                     user.District.String,
		Religion:                     user.Religion.String,
		MaritalStatus:                user.MaritalStatus.String,
		Profession:                   user.Profession.String,
		Citizenship:                  user.Citizenship.String,
		CardValidUntil:               cardValidUntil,
		Password:                     user.Password.String,
		PhoneNumber:                  user.PhoneNumber.String,
		PersonalIdentificationNumber: user.PersonalIdentificationNumber.String,
		EnableTouchID:                user.EnableTouchID,
		HashTouchID:                  user.HashTouchID.String,
		ProfilePhotoURL:              user.ProfilePhotoURL.String,
		IsVerified:                   user.IsVerified,
		CreatedAt:                    createdAt,
		UpdatedAt:                    updatedAt,
	}

	return userRes
}
