package user

// User is a registered person that can login to the
// application. Password is the encrypted password.
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
