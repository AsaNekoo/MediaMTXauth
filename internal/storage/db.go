package storage

import (
	"MediaMTXAuth/internal/passwords"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	Db *bolt.DB
}

type User struct {
	ID        int
	Name      string
	Hash      string
	Generated bool
}

const (
	usersBucket            = "users"
	userIndexBucket        = "user_index"
	namespacesBucket       = "namespaces"
	userDashSessionsBucket = "user_dash_sessions"
	streamSessionsBucket   = "stream_sessions"
)

var (
	ErrUsername = errors.New("username must be at least 3 characters long")

	ErrPassword = errors.New("passwords must be at least 8 characters long")
)

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func InitDB(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			usersBucket,
			userIndexBucket,
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

func (s *Storage) CreateUser(u *User, password string) error {
	trimmedName := strings.TrimSpace(u.Name)

	if trimmedName == "" || len(trimmedName) < 3 {
		return ErrUsername
	}

	if trim := strings.TrimSpace(password); trim == "" || len(trim) < 8 {
		return ErrPassword
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return err
	}

	return s.Db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(usersBucket))
		index := tx.Bucket([]byte(userIndexBucket))

		id, _ := users.NextSequence()
		u.ID = int(id)
		u.Hash = hash
		u.Name = trimmedName

		if existing := index.Get([]byte(trimmedName)); existing != nil {
			return errors.New("user already exists")
		}

		data, err := json.Marshal(u)
		if err != nil {
			return err
		}

		if err := users.Put(itob(u.ID), data); err != nil {
			return err
		}

		return index.Put([]byte(u.Name), itob(u.ID))
	})
}

func (s *Storage) GetUser(name string) (*User, error) {
	var u User
	err := s.Db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(usersBucket))
		index := tx.Bucket([]byte(userIndexBucket))

		// Get ID from name index
		idBytes := index.Get([]byte(strings.TrimSpace(name)))
		if idBytes == nil {
			return errors.New("user not found")
		}

		// Get user data by ID
		userData := users.Get(idBytes)
		if userData == nil {
			return errors.New("user not found")
		}

		return json.Unmarshal(userData, &u)
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Storage) VerifyUser(username, password string) (bool, error) {
	trimmedUsername := strings.TrimSpace(username)

	if trimmedUsername == "" || len(trimmedUsername) < 3 {
		return false, ErrUsername
	}

	if trim := strings.TrimSpace(password); trim == "" || len(trim) < 8 {
		return false, ErrPassword
	}

	var storedHash string
	err := s.Db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(usersBucket))
		index := tx.Bucket([]byte(userIndexBucket))

		id := index.Get([]byte(trimmedUsername))
		if id == nil {
			return errors.New("user not found")
		}

		userData := users.Get(id)
		if userData == nil {
			return errors.New("user not found")
		}

		var user User
		if err := json.Unmarshal(userData, &user); err != nil {
			return err
		}

		storedHash = user.Hash
		return nil
	})

	if err != nil {
		return false, err
	}

	return passwords.Verify(password, storedHash)
}
