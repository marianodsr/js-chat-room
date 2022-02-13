package users

import (
	"errors"
	"fmt"

	"github.com/marianodsr/jobsity-chat-room/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" db:"username" gorm:"unique"`
	Email    string `json:"email" db:"email" gorm:"unique"`
	Password string `json:"password" db:"password"`
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) createUser(user User) (*User, error) {

	_, err := s.repo.getUserByEmail(user.Email)
	fmt.Println(errors.Is(err, gorm.ErrRecordNotFound))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashedPassword, _ := utils.HashPassword(user.Password)

		user.Password = hashedPassword

		created, err := s.repo.createUser(user)
		if err != nil {
			return nil, err
		}
		return created, nil
	}
	return nil, fmt.Errorf("email or username already in use")
}

func (s *UserService) login(email string, password string) (*User, error) {
	dbUser, err := s.repo.getUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if ok := utils.CheckPasswordHash(password, dbUser.Password); !ok {
		return nil, fmt.Errorf("invalid email or password")
	}

	return dbUser, nil
}
