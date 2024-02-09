package exercise

import (
	"fmt"
)

// Exercise contains details of a single workout exercise
// an exercise is a node in a linked list, determining the order
type Exercise struct {
	Owner       string  `json:"owner"`
	Workout     int     `json:"workout"`
	Index       int     `json:"index"`
	Name        string  `json:"name" validate:"required"`
	Weight      float64 `json:"weight" validate:"required"`
	Repetitions int     `json:"repetitions" validate:"required"`
}

// String prints the Exercise is a human readable format implementing
// the String interface
func (e Exercise) String() string {
	return fmt.Sprintf("exercise %s: name %s, weight %.1f, reps %d",
		e.Key(),
		e.Name,
		e.Weight,
		e.Repetitions,
	)
}

// Key prints unique identifiable key
func (e Exercise) Key() string {
	return fmt.Sprintf("exercise %s/%d/%d created", e.Owner, e.Workout, e.Index)
}

// ByIndex implements sort.Interface, sorting exercises according to their index
type ByIndex []Exercise

func (xs ByIndex) Len() int           { return len(xs) }
func (xs ByIndex) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs ByIndex) Less(i, j int) bool { return xs[i].Index < xs[j].Index }
