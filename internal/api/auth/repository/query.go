package authRepository

const (
	queryCreateUser = `
INSERT INTO Users (id, phone_number, name, password, created_at)
VALUES (:id, :phone_number, :name, :password, :created_at)`

	queryGetById = `
SELECT id, email, name, national_identity_number, birth_place, birth_date, gender, 
       address, neighborhood_community_unit, village, district, religion, 
       marital_status, profession, citizenship, card_valid_until, password, 
       phone_number, personal_identification_number, enable_touch_id, hash_touch_id, 
       profile_photo_url, is_verified, created_at, updated_at, face_photo_url
FROM Users
    WHERE id = :id`

	queryGetByPhoneNumber = `
SELECT id, email, name, password, phone_number, is_verified
FROM Users
    WHERE phone_number = :phone_number`

	queryGetByEmail = `
SELECT id, email, name, national_identity_number, birth_place, birth_date, gender, address, neighborhood_community_unit, village, district, religion, marital_status, profession, citizenship, card_valid_until, password, phone_number, personal_identification_number, enable_touch_id, hash_touch_id, is_verified, face_photo_url, profile_photo_url, created_at, updated_at
FROM Users
    WHERE email = :email`

	queryUpdateUser = `
UPDATE Users 
SET name = :name, 
    national_identity_number = :national_identity_number, 
    birth_place = :birth_place, 
    birth_date = :birth_date, 
    gender = :gender, 
    address = :address, 
    neighborhood_community_unit = :neighborhood_community_unit, 
    village = :village, 
    district = :district, 
    religion = :religion, 
    marital_status = :marital_status, 
    profession = :profession, 
    citizenship = :citizenship, 
    card_valid_until = :card_valid_until, 
    is_verified = :is_verified,
    updated_at = :updated_at
WHERE id = :id`

	queryDeleteUser = `
DELETE FROM Users
WHERE id = :id`

	queryUpdateUserPassword = `
		UPDATE Users
SET password = :password
WHERE phone_number = :phone_number`

	queryUpdateEnableTouchID = `
		UPDATE Users
SET enable_touch_id = :enable_touch_id, hash_touch_id = :hash_touch_id
	WHERE id = :id`

	queryUpdateUserVerificationByPhoneNum = `
		UPDATE Users
SET is_verified = :is_verified, updated_at = :updated_at
WHERE phone_number = :phone_number`

	queryUpdateUserVerificationPIN = `
		UPDATE Users
SET is_verified = :is_verified, personal_identification_number = :personal_identification_number
WHERE phone_number = :phone_number`

	queryUpdateProfilePhoto = `
		UPDATE Users
		SET profile_photo_url = :profile_photo_url,
			updated_at = :updated_at
		WHERE id = :id`

	queryUpdateFacePhoto = `
		UPDATE Users
		SET face_photo_url = :face_photo_url,
			updated_at = :updated_at
		WHERE id = :id`
)
