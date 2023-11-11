package main

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// testExercises is a mock implementation for exercises
type testExercises struct {
	mocks map[string]Exercise
}

// newTestExercises inits the map with some example exercises
func newTestExercises(mocks ...Exercise) testExercises {
	xs := make(map[string]Exercise)
	for _, mock := range mocks {
		xs[mock.ID.String()] = mock
	}
	return testExercises{xs}
}

// ExerciseByID returns ErrExerciseNotFound if id is not used as key in mocks map
// otherwise it returns the mock Exercise
func (xs *testExercises) ExerciseByID(id uuid.UUID) (Exercise, error) {
	if _, ok := xs.mocks[id.String()]; !ok {
		return Exercise{}, ErrExerciseNotFound
	}
	return xs.mocks[id.String()], nil
}

// StoreExercise stores an Exercise in the mocks map, when ID already exists it
// overwrites the exercise
func (xs *testExercises) StoreExercise(e Exercise) (uuid.UUID, error) {
	xs.mocks[e.ID.String()] = e
	return e.ID, nil
}

// Exists returns true if id is in mocks
func (xs *testExercises) Exists(id uuid.UUID) bool {
	_, ok := xs.mocks[id.String()]
	return ok
}

func TestFetchSingleExerciseJSON(t *testing.T) {
	e1 := Exercise{uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6f"), "Benchpress", 90.5, 6, uuid.Nil, uuid.Nil}
	e1JSON, _ := json.Marshal(&e1)

	xs := newTestExercises(e1)

	cs := []struct {
		testname string
		given    string
		expected []byte
		err      error
	}{
		{"noId", "", emptyJSON, ErrNoIDProvided},
		{"InvalidIdFormat", "1234", emptyJSON, ErrInvalidIdFormat},
		{"IdNotFound", uuid.Nil.String(), emptyJSON, ErrExerciseNotFound},
		{"existingID", e1.ID.String(), e1JSON, nil},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			actual, err := FetchSingleExerciseJSON(&xs, c.given)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
			}
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestStoreExerciseJSON(t *testing.T) {
	e1 := Exercise{uuid.Nil, "New", 90.5, 6, uuid.Nil, uuid.Nil}
	e1JSON, _ := json.Marshal(&e1)

	e2 := Exercise{uuid.MustParse("5124a472-8315-4a0c-b9df-ed39bab3960c"), "Overwritten", 90.5, 6, uuid.Nil, uuid.Nil}
	e2JSON, _ := json.Marshal(&e2)

	e3 := Exercise{uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6f"), "Error", 90.5, 6, uuid.Nil, uuid.Nil}
	e3JSON, _ := json.Marshal(&e3)

	xs := newTestExercises(e2)

	cs := []struct {
		testname string
		given    []byte
		expected error
	}{
		{"invalidJSON", []byte("{\"invalid\":true"), ErrInvalidJSON},
		{"notAnExercise", []byte("{\"unknown\":1}"), ErrEmptyExercise},
		{"newExercise", e1JSON, nil},                         // no id provided
		{"existingExercise", e2JSON, nil},                    // existing id provided
		{"notExistingExercise", e3JSON, ErrExerciseNotFound}, // id provided, but does not exist
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			_, err := StoreExerciseJSON(&xs, c.given)
			assert.ErrorIs(t, err, c.expected)
		})
	}
}
