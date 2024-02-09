package workout

import "fmt"

// Workout is a collection of ordered exercises
// that should be completed in a single session
type Workout struct {
	Owner string `json:"owner"`
	Index int    `json:"index"`
	Name  string `json:"name" validate:"required"`
}

func (w Workout) String() string {
	return fmt.Sprintf("workout %s: %s", w.Key(), w.Name)
}

func (w Workout) Key() string {
	return fmt.Sprintf("%s/%d", w.Owner, w.Index)
}
