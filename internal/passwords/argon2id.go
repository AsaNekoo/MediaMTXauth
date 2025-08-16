package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/nothub/hashutils/encoding/b64"
	"github.com/nothub/hashutils/phc"
	"golang.org/x/crypto/argon2"
	"strconv"
)

var ErrNotEnoughArguments = errors.New("not enough arguments")

func Hash(password string) (string, error) {
	const (
		memory      = 19 * 1024
		iterations  = 2
		parallelism = 1
		saltLength  = 16
		keyLength   = 16
	)

	salt, err := generateSalt(saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	result := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash)

	return result, nil
}

func Verify(password, stored string) (bool, error) {
	p, err := phc.Parse(stored)

	if err != nil {
		return false, err
	}

	if p.Id() != "argon2id" {
		return false, nil
	}

	salt, err := b64.Decode(p.Salt())

	if err != nil {
		return false, err
	}

	hash, err := b64.Decode(p.Hash())

	if err != nil {
		return false, err
	}

	params := p.Params()
	var memory, iterations, threads uint64
	paramsParsed := 0

	for _, param := range params {
		switch param.K {
		case "t":
			iterations, err = strconv.ParseUint(param.V, 10, 32)
			paramsParsed++
		case "m":
			memory, err = strconv.ParseUint(param.V, 10, 32)
			paramsParsed++
		case "p":
			threads, err = strconv.ParseUint(param.V, 10, 8)
			paramsParsed++
		}

		if err != nil {
			return false, err
		}
	}

	if paramsParsed < 3 {
		return false, ErrNotEnoughArguments
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(iterations),
		uint32(memory),
		uint8(threads),
		uint32(len(hash)),
	)

	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

func generateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
