package bcrypt

import "golang.org/x/crypto/bcrypt"

type IBcrypt interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashPassword string, password string) error
}

type bcryptService struct {
	cost int
}

func New() IBcrypt {
	return &bcryptService{
		cost: bcrypt.DefaultCost,
	}
}

func NewWithCost(cost int) IBcrypt {
	return &bcryptService{
		cost: cost,
	}
}

func (b *bcryptService) HashPassword(password string) (string, error) {
	pass := []byte(password)
	result, err := bcrypt.GenerateFromPassword(pass, b.cost)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (b *bcryptService) ComparePassword(hashPassword string, password string) error {
	pass := []byte(password)
	hashPass := []byte(hashPassword)
	return bcrypt.CompareHashAndPassword(hashPass, pass)
}
