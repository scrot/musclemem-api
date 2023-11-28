package workout

// Workout is a collection of ordered exercises
// that should be completed in a single session
type Workout struct {
	ID    int    `json:"workout-id" validate:"required,gt=0"`
	Owner int    `json:"owner" validate:"required,gt=0"`
	Name  string `json:"name" validate:"required"`
}
