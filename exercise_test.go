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
		{"withID", e1.ID.String(), e1JSON, nil},
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
