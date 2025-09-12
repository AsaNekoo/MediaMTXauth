package services

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/passwords"
	"MediaMTXAuth/internal/storage"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"
)

const DefaultAdminUsername = "admin"
const DefaultAdminPassword = "admin"

var (
	ErrShortUsername = errors.New("username must be at least 3 characters long")
	ErrShortPassword = errors.New("password must be at least 8 characters long")
)

type userService struct {
	storage storage.Storage
}

func NewUserService(storage storage.Storage) internal.UserService {
	return &userService{storage}
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return ErrShortUsername
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrShortPassword
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

	return s.create(username, password, isAdmin)
}

func (s *userService) create(username, password string, isAdmin bool) (*internal.User, error) {
	existingUser, err := s.storage.GetUser(username)

	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, internal.ErrUserAlreadyExists
	}

	hash, err := passwords.Hash(password)

	if err != nil {
		return nil, err
	}

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
	_, err := s.create(DefaultAdminUsername, DefaultAdminPassword, true)

	if err == nil {
		return DefaultAdminPassword, nil
	}

	if errors.Is(err, internal.ErrUserAlreadyExists) {
		return "", nil
	}

	return "", err
}

func (s *userService) Get(username string) (*internal.User, error) {
	return s.storage.GetUser(username)
}

func (s *userService) Delete(username string) error {
	return s.storage.DeleteUser(username)
}

func (s *userService) ChangePassword(username, password string) error {
	user, err := s.storage.GetUser(username)

	if err != nil {
		return err
	}

	if user == nil {
		return internal.ErrUserNotFound
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

	if user == nil {
		return "", internal.ErrUserNotFound
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

	if user == nil {
		return "", internal.ErrUserNotFound
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

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, internal.ErrUserNotFound
	}

	storedHash := user.Password.Hash
	p, err := passwords.Verify(password, storedHash)
	if err != nil {
		return nil, err
	}

	if !p {
		return nil, internal.ErrWrongPassword
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

func (s *userService) Logout(username string) (*internal.User, error) {
	user, err := s.storage.GetUser(username)

	if err != nil || user == nil {
		return nil, err
	}

	user.Session = internal.UserSession{}

	err = s.storage.SetUser(*user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) VerifySession(username, sessionkey string) (bool, error) {
	user, err := s.storage.GetUser(username)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, internal.ErrUserNotFound
	}

	if user.Session.ID == 0 || user.Session.Expiration.Before(time.Now()) {
		return false, nil
	}

	idStr := fmt.Sprintf("%d", user.Session.ID)
	if subtle.ConstantTimeCompare([]byte(idStr), []byte(sessionkey)) == 1 {
		return true, nil
	}
	return false, nil
}
