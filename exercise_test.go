package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchSingleExerciseJSON(t *testing.T) {
	cs := []struct {
		testname string
		given    string
		expected string
		err      error
	}{
		{"noId", "", "{}", ErrNoIDProvided},
		{"InvalidIdFormat", "1234", "{}", ErrInvalidIdFormat},
		// {"IdNotFound", uuid.Nil.String(), "{}", ErrExerciseNotFound},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			actual, err := FetchSingleExerciseJSON(c.given)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
			}
			assert.Equal(t, c.expected, actual)
		})
	}
}
