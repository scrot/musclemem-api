package workout

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Workout is a collection of ordered exercises
// that should be completed in a single session
type Workout struct {
	Owner string `json:"owner"`
	Index int    `json:"index"`
	Name  string `json:"name" validate:"required"`
}

func (w Workout) String() string {
	return fmt.Sprintf("workout %s: %s", w.Ref(), w.Name)
}

func (w Workout) Ref() WorkoutRef {
	return WorkoutRef{w.Owner, w.Index}
}

// WorkoutRef represents the unique key that references an Workout
type WorkoutRef struct {
	Username     string
	WorkoutIndex int
}

func (wr WorkoutRef) String() string {
	return fmt.Sprintf("%s/%d", wr.Username, wr.WorkoutIndex)
}

func ParseRef(s string) (WorkoutRef, error) {
	ss := strings.Split(s, "/")

	if len(ss) != 2 || ss[0] == "" || ss[1] == "" {
		return WorkoutRef{}, errors.New("invalid ref expected {username}/{workout-index}")
	}

	wi, err := strconv.Atoi(ss[1])
	if err != nil {
		return WorkoutRef{}, err
	}

	return WorkoutRef{Username: ss[0], WorkoutIndex: wi}, nil
}
