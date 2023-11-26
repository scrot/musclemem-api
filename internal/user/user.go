package user

type User struct {
	ID       int
	Email    string
	Password string
}

type Registerer interface {
	Register(email, password string) bool
}
