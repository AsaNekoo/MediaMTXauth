package storage

import (
	"MediaMTXAuth/internal/api"
	"errors"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	Db *bolt.DB
}

const (
	usersBucket            = "users"
	namespacesBucket       = "namespaces"
	userDashSessionsBucket = "user_dash_sessions"
	streamSessionsBucket   = "stream_sessions"
)

var (
	ErrUsername = errors.New("username must be at least 3 characters long")

	ErrPassword = errors.New("password must be at least 8 characters long")
)

func InitDB(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			usersBucket,
			namespacesBucket,
			userDashSessionsBucket,
			streamSessionsBucket,
		}
		for _, b := range buckets {
			if _, e := tx.CreateBucketIfNotExists([]byte(b)); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return &Storage{Db: db}, nil
}

func (s *Storage) Close() error {
	return s.Db.Close()
}

func (s *Storage) CreateUser(username, password string) error {

	if trim := strings.TrimSpace(username); trim == "" || len(trim) < 3 {
		return ErrUsername
	}

	if trim := strings.TrimSpace(password); trim == "" || len(trim) < 8 {
		return ErrPassword
	}

	hash, err := api.HashPassword(password)
	if err != nil {
		return err
	}

	return s.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		return b.Put([]byte(strings.TrimSpace(username)), []byte(hash))
	})
}

func (s *Storage) VerifyUser(username, password string) (bool, error) {

	if trim := strings.TrimSpace(username); trim == "" || len(trim) < 3 {
		return false, ErrUsername
	}

	if trim := strings.TrimSpace(password); trim == "" || len(trim) < 8 {
		return false, ErrPassword
	}

	var stored string
	err := s.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		val := b.Get([]byte(strings.TrimSpace(username)))
		if val == nil {
			return errors.New("user not found")
		}
		stored = string(val)
		return nil
	})
	if err != nil {
		return false, err
	}
	return api.VerifyPassword(password, stored)
}
