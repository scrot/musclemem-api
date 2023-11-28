package workout

import "github.com/scrot/musclemem-api/internal/exercise"

// Workout is a collection of ordered exercises
// that should be completed in a single session
type Workout struct {
	ID    int
	Owner int
	Name  string `json:"name" validate:"required"`
}

// Exercises returns all exercises associated with a workout id
// Empty list will be returned if there a no exercises or an error
func (w Workout) Exercises(ws Workouts) []exercise.Exercise {
	xs, err := ws.WorkoutExercises(w.ID)
	if err != nil {
		return []exercise.Exercise{}
	}
	return xs
}

type Service struct {
	ws Workouts
}
