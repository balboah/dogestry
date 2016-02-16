package remote

import (
	"strings"
)

type ID string

func NewID(hash string) ID {
	return ID(trimHashPrefix(hash))
}

// trimHashPrefix removes hashing definition that we are not interested in here.
// Also makes sure as to not break backwards compatibility with older versions
// where Docker did not specify the hash function.
func trimHashPrefix(s string) string {
	return strings.TrimPrefix(string(s), "sha256:")
}

func (id ID) Short() ID {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return ID(id[:shortLen])
}

func (id ID) String() string {
	return string(id)
}
