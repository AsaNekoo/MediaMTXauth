package bolt

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/storage"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"time"
)

var usersBucket = []byte("users")
var namespacesBucket = []byte("namespaces")
var buckets = [][]byte{usersBucket, namespacesBucket}

type boltStorage struct {
	DB *bolt.DB
}

func New(path string) (storage.Storage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, err
	}

	return &boltStorage{DB: db}, nil
}

func (s *boltStorage) Close() error {
	return s.DB.Close()
}

func (s *boltStorage) Init() error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *boltStorage) SetUser(u internal.User) error {
	return set(s.DB, usersBucket, u)
}

func (s *boltStorage) GetUser(name string) (*internal.User, error) {
	return get[internal.User](s.DB, usersBucket, name)
}

func (s *boltStorage) SetNamespace(u internal.Namespace) error {
	return set(s.DB, namespacesBucket, u)
}

func (s *boltStorage) GetNamespace(name string) (*internal.Namespace, error) {
	return get[internal.Namespace](s.DB, namespacesBucket, name)
}

func set[T internal.WithID](db *bolt.DB, bucket []byte, v T) error {
	data, err := json.Marshal(&v)

	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(v.GetID()), data)
	})
}

func get[T any](db *bolt.DB, bucket []byte, name string) (u *T, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucket).Get([]byte(name))

		if data == nil {
			return nil
		}

		var stored T

		if err := json.Unmarshal(data, &stored); err != nil {
			return err
		}

		u = &stored

		return nil
	})

	return
}
