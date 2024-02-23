package exercise

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Exercise contains details of a single workout exercise
// an exercise is a node in a linked list, determining the order
type Exercise struct {
	Owner       string  `json:"owner"`
	Workout     int     `json:"workout"`
	Index       int     `json:"index"`
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Repetitions int     `json:"repetitions"`
}

// String prints the Exercise is a human readable format implementing
// the String interface
func (e Exercise) String() string {
	return fmt.Sprintf("exercise %s: name %s, weight %.1f, reps %d",
		e.Ref(),
		e.Name,
		e.Weight,
		e.Repetitions,
	)
}

// Key prints unique identifiable key
func (e Exercise) Ref() ExerciseRef {
	return ExerciseRef{e.Owner, e.Workout, e.Index}
}

// ExerciseRef represents the unique key that references an Exercise
type ExerciseRef struct {
	Username      string
	WorkoutIndex  int
	ExerciseIndex int
}

func (er ExerciseRef) String() string {
	return fmt.Sprintf("%s/%d/%d", er.Username, er.WorkoutIndex, er.ExerciseIndex)
}

func ParseRef(s string) (ExerciseRef, error) {
	ss := strings.Split(s, "/")

	if len(ss) != 3 || ss[0] == "" || ss[1] == "" || ss[2] == "" {
		return ExerciseRef{}, errors.New("invalid ref expected {username}/{workout-index}/{exercise-index}")
	}

	wi, err := strconv.Atoi(ss[1])
	if err != nil {
		return ExerciseRef{}, err
	}

	return ExerciseRef{Username: ss[0], WorkoutIndex: wi}, nil
}

// With is used to represent the target exercise used to swap exercises
type With struct {
	Exercise ExerciseRef
}

// ByIndex implements sort.Interface, sorting exercises according to their index
type ByIndex []Exercise

func (xs ByIndex) Len() int           { return len(xs) }
func (xs ByIndex) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs ByIndex) Less(i, j int) bool { return xs[i].Index < xs[j].Index }
