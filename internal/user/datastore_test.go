package user_test

import (
	"errors"
	"testing"

	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/user"
)

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	us := internal.NewMockSqliteDatastore(t)

	cs := []struct {
		name          string
		email         string
		password      string
		expectedNewID bool
		expectedErr   error
	}{
		{"ErrorOnMissingEmail", "", "passwd", false, user.ErrMissingFields},
		{"ErrorOnMissingPassword", "e@mail.com", "", false, user.ErrMissingFields},
		{"ErrorOnInvalidEmailMissing", "@email.com", "passwd", false, user.ErrInvalidValue},
		{"ErrorOnInvalidEmailSymbol@", "at@@email.com", "passwd", false, user.ErrInvalidValue},
		{"NewIDOnValidLocalEmail", "a-t@localdomain", "passwd", true, nil},
		{"NewIDOnValidPublicEmail", "e@gmail.com", "passwd", true, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			id, err := us.Users.Register(c.email, c.password)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			isNewID := id > 0
			if isNewID != c.expectedNewID {
				t.Errorf("expected new id to be %t but got %t", c.expectedNewID, isNewID)
			}
		})
	}
}
