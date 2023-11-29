package user

import (
	"fmt"

	"github.com/scrot/musclemem-api/internal/workout"
)

// User is a registered person that can login to the
// application. Password is the encrypted password.
type User struct {
	ID       int
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Workouts returns all workouts belonging to that user
func (u User) Workouts(us Retreiver) ([]workout.Workout, error) {
	return []workout.Workout{}, fmt.Errorf("not implemented")
}
