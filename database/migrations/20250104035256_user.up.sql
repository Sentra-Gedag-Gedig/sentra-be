CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(26) PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    national_identity_number VARCHAR(255) UNIQUE,
    birth_place VARCHAR(255),
    birth_date DATE,
    gender VARCHAR(26),
    address VARCHAR(255),
    neighborhood_community_unit VARCHAR(255),
    village VARCHAR(255),
    district VARCHAR(255),
    religion VARCHAR(255),
    marital_status VARCHAR(255),
    profession VARCHAR(255),
    citizenship VARCHAR(255),
    card_valid_until DATE,
    password VARCHAR(255),
    phone_number VARCHAR(255) UNIQUE,
    personal_identification_number VARCHAR(255),
    enable_touch_id BOOLEAN DEFAULT false,
    hash_touch_id VARCHAR(255),
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    profile_photo_url VARCHAR(255) DEFAULT NULL,
    face_photo_url VARCHAR(255) DEFAULT NULL
    );
