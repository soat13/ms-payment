package domain

import (
	"errors"
	"strings"
)

type Link string

var ErrInvalidLink = errors.New("invalid link")

func (link Link) Validate() error {
	if !strings.HasPrefix(string(link), "https://") {
		return ErrInvalidLink
	}

	return nil
}
