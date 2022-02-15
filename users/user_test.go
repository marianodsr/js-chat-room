package users

import (
	"fmt"
	"testing"

	"github.com/marianodsr/jobsity-chat-room/utils"
	"gorm.io/gorm"
)

type mockRepo struct {
	shouldGetUserByEmailErr bool
	shouldCreateUserErr     bool
	err                     error
	getUserByEmailResp      *User
}

func (m *mockRepo) getUserByEmail(email string) (*User, error) {
	if m.shouldGetUserByEmailErr {
		return nil, m.err
	}
	return m.getUserByEmailResp, nil
}

func (m *mockRepo) createUser(user User) (*User, error) {
	if m.shouldCreateUserErr {
		return nil, fmt.Errorf("mocked error")
	}
	return &user, nil
}

func NewMockService() *UserService {
	return &UserService{
		repo: &mockRepo{
			shouldGetUserByEmailErr: false,
			shouldCreateUserErr:     false,
		},
	}
}

func TestCreateUser(t *testing.T) {

	service := NewMockService()
	table := []struct {
		testName                string
		username                string
		password                string
		email                   string
		shouldError             bool
		shouldGetUserByEmailErr bool
		shouldCreateUserErr     bool
		mockGetUserByEmailErr   error
	}{
		{
			testName:                "happy_path",
			username:                "user123",
			password:                "123",
			email:                   "email@email.com",
			shouldError:             false,
			shouldGetUserByEmailErr: true,
			shouldCreateUserErr:     false,
			mockGetUserByEmailErr:   gorm.ErrRecordNotFound,
		},
		{
			testName:                "getUserByEmail_dbError",
			username:                "user123",
			password:                "123",
			email:                   "email@email.com",
			shouldError:             true,
			shouldGetUserByEmailErr: true,
			shouldCreateUserErr:     false,
			mockGetUserByEmailErr:   fmt.Errorf("just a regular db error"),
		},
		{
			testName:                "getUserByEmail_ItFoundsAnExistingUser",
			username:                "user123",
			password:                "123",
			email:                   "email@email.com",
			shouldError:             true,
			shouldGetUserByEmailErr: false,
			shouldCreateUserErr:     false,
			mockGetUserByEmailErr:   nil,
		},
		{
			testName:                "createUser_dbError",
			username:                "user123",
			password:                "123",
			email:                   "email@email.com",
			shouldError:             true,
			shouldGetUserByEmailErr: true,
			shouldCreateUserErr:     true,
			mockGetUserByEmailErr:   gorm.ErrRecordNotFound,
		},
	}

	repo, _ := service.repo.(*mockRepo)

	for _, tt := range table {
		t.Run(tt.testName, func(t *testing.T) {
			repo.shouldGetUserByEmailErr = false
			repo.shouldCreateUserErr = false
			if tt.shouldGetUserByEmailErr {
				repo.shouldGetUserByEmailErr = true
				repo.err = tt.mockGetUserByEmailErr
			}
			if tt.shouldCreateUserErr {
				repo.shouldCreateUserErr = true
			}

			user, err := service.createUser(User{
				Model:    gorm.Model{},
				Email:    tt.email,
				Password: tt.password,
				Username: tt.username,
			})
			if err == nil && tt.shouldError {
				t.Errorf("wanted error but got nil")
			}

			if !tt.shouldError {
				if user.Email != tt.email {
					t.Errorf("actual is different than expected\nActual: %+v\nExpected: %+v\n", user, User{gorm.Model{}, tt.email, tt.username, tt.password})
				}
			}

		})
	}
}

func TestLogin(t *testing.T) {

	service := NewMockService()
	repo, _ := service.repo.(*mockRepo)

	table := []struct {
		testName                string
		email                   string
		password                string
		shouldError             bool
		shouldGetUserByEmailErr bool
		getUserByEmailErr       error
		getUserByEmailResp      *User
	}{
		{
			testName:                "happy_path",
			email:                   "email@email.com",
			password:                "123",
			shouldError:             false,
			shouldGetUserByEmailErr: false,
			getUserByEmailErr:       nil,
			getUserByEmailResp: &User{
				Email:    "email@email.com",
				Password: "123",
			},
		},
		{
			testName:                "getUserByEmail_DBError",
			email:                   "email@email.com",
			password:                "123",
			shouldError:             true,
			shouldGetUserByEmailErr: true,
			getUserByEmailErr:       fmt.Errorf("Im a mocked error"),
			getUserByEmailResp: &User{
				Email:    "email@email.com",
				Password: "123",
			},
		},
		{
			testName:                "wrongPassword_Error",
			email:                   "email@email.com",
			password:                "123xasda",
			shouldError:             true,
			shouldGetUserByEmailErr: false,
			getUserByEmailErr:       fmt.Errorf("Im a mocked error"),
			getUserByEmailResp: &User{
				Email:    "email@email.com",
				Password: "123",
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.getUserByEmailResp != nil {
				hashed, _ := utils.HashPassword(tt.getUserByEmailResp.Password)
				tt.getUserByEmailResp.Password = hashed
			}

			repo.getUserByEmailResp = tt.getUserByEmailResp
			if tt.shouldGetUserByEmailErr {
				repo.shouldGetUserByEmailErr = true
				repo.err = tt.getUserByEmailErr
			}

			user, err := service.login(tt.email, tt.password)

			if tt.shouldError && err == nil {
				t.Errorf("wanted err but got nil")
			}

			if !tt.shouldError {
				if user.Email != tt.email {
					t.Errorf("expected equal emails but got:\nActual:\n  Email:%s\nExpected:\n  Email:%s", tt.email, user.Email)
				}
			}

		})
	}
}
