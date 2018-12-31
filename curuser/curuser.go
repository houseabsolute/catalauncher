package curuser

import (
	"fmt"
	"os/user"
)

type User struct {
	*user.User
}

func New() (*User, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("Could not get the current user: %s", err)
	}
	return &User{u}, nil
}
