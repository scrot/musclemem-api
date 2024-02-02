package exercise

import "fmt"

// Exercise contains details of a single workout exercise
// an exercise is a node in a linked list, determining the order
type Exercise struct {
	ID          int
	Workout     int     `json:"workout" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Weight      float64 `json:"weight" validate:"required"`
	Repetitions int     `json:"repetitions" validate:"required"`
	NextID      int
	PreviousID  int
}

// String prints the Exercise is a human readable format implementing
// the String interface
func (e Exercise) String() string {
	return fmt.Sprintf("exercise %s (weight %.1f, reps %d)",
		e.Name,
		e.Weight,
		e.Repetitions,
	)
}

// Next returns next Exercise node in linked list
// returns empty exercise if already at the tail
func (e Exercise) Next(r Retreiver) Exercise {
	x, err := r.WithID(e.NextID)
	if err != nil {
		return Exercise{}
	}
	return x
}

// Previous returns previous Exercise node in linked list
// returns empty exercise if already at the head
func (e Exercise) Previous(r Retreiver) Exercise {
	x, err := r.WithID(e.PreviousID)
	if err != nil {
		return Exercise{}
	}
	return x
}
