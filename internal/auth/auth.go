package auth

import (
	"MediaMTXAuth/internal"
	"errors"
	"fmt"
)

type Auth struct {
	UserService      internal.UserService
	NamespaceService internal.NamespaceService
}

func New(userService internal.UserService, namespaceService internal.NamespaceService) *Auth {
	return &Auth{
		UserService:      userService,
		NamespaceService: namespaceService,
	}
}

func (a *Auth) Validate(namespace, userName, streamKey, action string) error {
	_, err := a.NamespaceService.Get(namespace)

	if errors.Is(err, internal.ErrNamespaceNotFound) {
		return fmt.Errorf("%w: %w", ErrAuthError, err)
	} else if err != nil {
		return err
	}

	user, err := a.UserService.Get(userName)

	if err != nil || user == nil {
		return fmt.Errorf("%w: %w", ErrAuthError, internal.ErrUserNotFound)
	}

	if user.Namespace != "" && user.Namespace != namespace {
		return fmt.Errorf("%w: %s", ErrAuthError, "user is not allowed in name")
	}

	if action == "publish" && user.StreamKey != streamKey {
		return fmt.Errorf("%w: %s", ErrAuthError, "invalid stream key for user")
	}

	return nil
}
