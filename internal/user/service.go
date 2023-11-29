package user

type User struct {
	ID       int
	Email    string `json:"email"`
	Password string `json:"password"`
}
