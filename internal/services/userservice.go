package services

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/passwords"
	"MediaMTXAuth/internal/storage"
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"time"
)

var (
	ErrUsername  = errors.New("username must be at least 3 characters long")
	ErrPassword  = errors.New("password must be at least 8 characters long")
	ErrIncorrect = errors.New("incorrect")
)

type userService struct {
	storage storage.Storage
}

func NewUserService() *userService {
	return &userService{}
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return ErrUsername
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPassword
	}
	return nil
}

func (s *userService) Create(username, password string, isAdmin bool) (*internal.User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	existingUser, err := s.storage.GetUser(username)
	if err == nil && existingUser != nil {
	}

	hash, err := passwords.Hash(password)

	userPassword := internal.UserPassword{
		Hash:        hash,
		IsGenerated: true,
	}

	user := internal.User{
		Name:      username,
		StreamKey: rand.Text(),
		IsAdmin:   isAdmin,
		Password:  userPassword,
	}

	err = s.storage.SetUser(user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) CreateDefaultAdminUser() (string, error) {

	const (
		password string = "admin"
		username string = "admin"
		isAdmin  bool   = true
	)

	existingUser, err := s.storage.GetUser(username)
	if err == nil && existingUser != nil {
	}

	hash, err := passwords.Hash(password)

	userPassword := internal.UserPassword{
		Hash:        hash,
		IsGenerated: true,
	}

	user := internal.User{
		Name:     username,
		IsAdmin:  isAdmin,
		Password: userPassword,
	}

	err = s.storage.SetUser(user)
	if err != nil {
		return "", err
	}

	result := password + username
	return result, nil
}

func (s *userService) Get(username string) (*internal.User, error) {
	var user *internal.User
	var err error

	user, err = s.storage.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(username string) error {
	var err error

	err = s.storage.DeleteUser(username)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) ChangePassword(username, password string) error {
	user, err := s.storage.GetUser(username)
	if err != nil {
		return err
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return err
	}

	user.Password = internal.UserPassword{
		Hash:        hash,
		IsGenerated: false,
	}

	return s.storage.SetUser(*user)
}

func (s *userService) ResetPassword(username string) (string, error) {
	user, err := s.storage.GetUser(username)
	if err != nil {
		return "", err
	}

	generated := rand.Text()

	hash, err := passwords.Hash(generated)
	if err != nil {
		return "", err
	}

	user.Password = internal.UserPassword{
		Hash:        hash,
		IsGenerated: true,
	}

	err = s.storage.SetUser(*user)
	if err != nil {
		return "", err
	}
	return generated, nil
}

func (s *userService) ResetStreamKey(username string) (string, error) {
	user, err := s.storage.GetUser(username)
	if err != nil {
		return "", err
	}

	generated := rand.Text()
	user.StreamKey = generated

	err = s.storage.SetUser(*user)
	if err != nil {
		return "", err
	}
	return generated, nil
}

func (s *userService) Login(username, password string) (*internal.User, error) {
	user, err := s.storage.GetUser(username)

	storedHash := user.Password.Hash
	p, err := passwords.Verify(password, storedHash)
	if p != true {
		return nil, ErrIncorrect
	}
	if err != nil {
		return nil, err
	}

	randomID, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	user.Session = internal.UserSession{
		ID:         uint64(randomID.Int64()),
		Expiration: time.Now().Add(time.Minute * 15),
	}

	err = s.storage.SetUser(*user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Logout(username string) error {
	user, err := s.storage.GetUser(username)
	if err != nil {
		return err
	}

	user.Session = internal.UserSession{}

	return s.storage.SetUser(*user)
}
