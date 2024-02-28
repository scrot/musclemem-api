package user

import (
	"errors"
	"strings"
	"testing"

	"github.com/scrot/musclemem-api/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser(t *testing.T) {
	validUser := User{"valid", "test@gmail.com", "-"}

	cs := []struct {
		name        string
		username    string
		email       string
		password    string
		want        User
		wantErr     error
		wantPassErr bool
	}{
		{"validUser", "valid", "test@gmail.com", "secret", validUser, nil, false},
		{"missingUser", "", "test@gmail.com", "secret", User{}, ErrEmptyField, true},
	}
	users, flush := mockUserStore(t)
	defer flush()

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			got, err := users.New(c.username, c.email, c.password)
			if !errors.Is(err, c.wantErr) {
				t.Error(err)
			}

			if strings.Compare(got.Username, c.want.Username) != 0 {
				t.Errorf("want %s but got %s", c.want.Username, got.Username)
			}

			if strings.Compare(got.Email, c.want.Email) != 0 {
				t.Errorf("want %s but got %s", c.want.Email, got.Email)
			}

			err = bcrypt.CompareHashAndPassword([]byte(got.Password), []byte(c.password))
			if (err != nil) != c.wantPassErr {
				t.Errorf("password does not match hash")
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	users, flush := mockUserStore(t)
	defer flush()

	validUser, err := users.New("user", "test@gmail.com", "secret")
	if err != nil {
		t.Fatal(err)
	}

	cs := []struct {
		name     string
		username string
		want     User
		wantErr  error
	}{
		{"validUser", "user", validUser, nil},
		{"missingUser", "", User{}, ErrEmptyField},
		{"unknownUser", "unknown", User{}, ErrUnknownUser},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			got, err := users.ByUsername(c.username)
			if !errors.Is(err, c.wantErr) {
				t.Error(err)
			}

			if strings.Compare(got.Username, c.want.Username) != 0 {
				t.Errorf("want %s but got %s", c.want.Username, got.Username)
			}

			if strings.Compare(got.Email, c.want.Email) != 0 {
				t.Errorf("want %s but got %s", c.want.Email, got.Email)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	users, flush := mockUserStore(t)
	defer flush()

	validUser, err := users.New("user", "test@gmail.com", "secret")
	if err != nil {
		t.Fatal(err)
	}

	cs := []struct {
		name     string
		username string
		password string
		want     User
		wantErr  error
	}{
		{"validUser", "user", "secret", validUser, nil},
		{"unauthenticatedUser", "user", "wrong", User{}, ErrWrongPassword},
		{"missingUser", "", "secret", User{}, ErrEmptyField},
		{"unknownUser", "unknown", "secret", User{}, ErrUnknownUser},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			got, err := users.Authenticate(c.username, c.password)
			if !errors.Is(err, c.wantErr) {
				t.Error(err)
			}

			if strings.Compare(got.Username, c.want.Username) != 0 {
				t.Errorf("want %s but got %s", c.want.Username, got.Username)
			}

			if strings.Compare(got.Email, c.want.Email) != 0 {
				t.Errorf("want %s but got %s", c.want.Email, got.Email)
			}

			if strings.Compare(got.Password, c.want.Password) != 0 {
				t.Errorf("want %s but got %s", c.want.Password, got.Password)
			}
		})
	}
}

func mockUserStore(t *testing.T) (UserStore, func()) {
	t.Helper()

	config := storage.DatastoreConfig{
		DatabaseURL:   "file://test.db?cache=shared&mode=memory",
		MigrationPath: "migrations",
		Overwrite:     false,
	}

	store, err := storage.NewSqlDatastore(config)
	if err != nil {
		t.Fatal(err)
	}

	flush := func() {
		if err := store.Close(); err != nil {
			t.Fatal()
		}
	}

	return NewSQLUserStore(store), flush
}
