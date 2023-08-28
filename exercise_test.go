package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type exerciseTestStorage struct {
	exercises map[int]Exercise
}

func newExerciseTestStorage() *exerciseTestStorage {
	e1 := Exercise{
		ID:          1,
		Name:        "Barbell Bench Press",
		Weight:      100,
		Repetitions: 5,
	}

	e2 := Exercise{
		ID:          2,
		Name:        "Larsen Press",
		Weight:      60,
		Repetitions: 10,
	}

	e3 := Exercise{
		ID:          3,
		Name:        "Arnold Press",
		Weight:      18,
		Repetitions: 8,
	}

	e1.Next = &e2
	e2.Next = &e3

	e3.Previous = &e2
	e2.Previous = &e1

	return &exerciseTestStorage{
		exercises: map[int]Exercise{
			e1.ID: e1,
			e2.ID: e2,
			e3.ID: e3,
		},
	}
}

func (s *exerciseTestStorage) ExerciseByID(id int) (Exercise, error) {
	if e, ok := s.exercises[id]; ok {
		return e, nil
	}
	return Exercise{}, nil
}

func TestSingleExerciseResponse(t *testing.T) {
	s := NewServer(ServerConfig{}, slog.Default(), newExerciseTestStorage())

	r := httptest.NewRequest(http.MethodGet, "/exercise/1", nil)
	w := httptest.NewRecorder()

	s.HandleExerciseDetails(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
