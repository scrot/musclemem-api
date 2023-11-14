package musclememapi_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	musclememapi "github.com/scrot/musclemem-api"
	"github.com/stretchr/testify/assert"
)

// testExercises is a mock implementation for exercises
type testExercises struct {
	mocks map[string]musclememapi.Exercise
}

// newTestExercises inits the map with some example exercises
func newTestExercises(mocks ...musclememapi.Exercise) testExercises {
	xs := make(map[string]musclememapi.Exercise)
	for _, mock := range mocks {
		xs[mock.ID.String()] = mock
	}
	return testExercises{xs}
}

// ExerciseByID returns ErrExerciseNotFound if id is not used as key in mocks map
// otherwise it returns the mock Exercise
func (xs *testExercises) ExerciseByID(id uuid.UUID) (musclememapi.Exercise, error) {
	if _, ok := xs.mocks[id.String()]; !ok {
		return musclememapi.Exercise{}, musclememapi.ErrExerciseNotFound
	}
	return xs.mocks[id.String()], nil
}

// StoreExercise stores an Exercise in the mocks map, when ID already exists it
// overwrites the exercise
func (xs *testExercises) StoreExercise(e musclememapi.Exercise) (musclememapi.Exercise, error) {
	xs.mocks[e.ID.String()] = e
	return xs.mocks[e.ID.String()], nil
}

// Exists returns true if id is in mocks
func (xs *testExercises) Exists(id uuid.UUID) bool {
	_, ok := xs.mocks[id.String()]
	return ok
}

func newTestSqliteDatastore(t *testing.T, xs ...musclememapi.Exercise) *musclememapi.SqliteDatastore {
	t.Helper()
	db, err := musclememapi.NewSqliteDatastore("file://"+t.TempDir(), true)
	if err != nil {
		t.Error(err)
	}
	return db
}

func TestExerciseExists(t *testing.T) {
	xs := newTestSqliteDatastore(t)
	xs.Exec(`
    INSERT INTO 'exercises'
    VALUES('6c255201-30dc-40c5-a016-b5374a7c4d6f', 'Test', 100, 10, NULL, NULL);
    `)

	valid := uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6f")
	invalid := uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6l")

	cs := []struct {
		testname string
		given    uuid.UUID
		expected bool
		err      error
	}{
		{"existingExercise", valid, true, nil},
		{"nonExistingExercise", invalid, false, nil},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			got := xs.Exists(valid)
			assert.Equal(t, c.expected, got)
		})
	}
}

func TestFetchSingleExerciseJSON(t *testing.T) {
	e1 := musclememapi.Exercise{uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6f"), "Benchpress", 90.5, 6, nil, nil}
	e1JSON, _ := json.Marshal(&e1)

	xs := newTestExercises(e1)

	cs := []struct {
		testname string
		given    string
		expected []byte
		err      error
	}{
		{"ErrorOnNoId", "", musclememapi.EmptyJSON, musclememapi.ErrNoIDProvided},
		{"ErrorOnInvalidIdFormat", "1234", musclememapi.EmptyJSON, musclememapi.ErrInvalidIdFormat},
		{"ErrorOnIdNotFound", uuid.Nil.String(), musclememapi.EmptyJSON, musclememapi.ErrExerciseNotFound},
		{"EqExerciseOnExistingID", e1.ID.String(), e1JSON, nil},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			actual, err := musclememapi.FetchSingleExerciseJSON(&xs, c.given)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
			}
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestStoreSingleExerciseJSON(t *testing.T) {
	// empty repository
	xs := newTestSqliteDatastore(t)
	defer xs.Close()

	// test exercises
	empty := musclememapi.Exercise{}

	e1 := musclememapi.Exercise{uuid.Nil, "New", 90.5, 6, nil, nil}
	e1JSON, _ := json.Marshal(&e1)

	e2 := musclememapi.Exercise{uuid.MustParse("5124a472-8315-4a0c-b9df-ed39bab3960c"), "Overwritten", 90.5, 6, nil, nil}
	e2JSON, _ := json.Marshal(&e2)

	e3 := musclememapi.Exercise{uuid.MustParse("6c255201-30dc-40c5-a016-b5374a7c4d6f"), "Error", 90.5, 6, nil, nil}
	e3JSON, _ := json.Marshal(&e3)

	// insert e2

	cs := []struct {
		testname string
		given    []byte
		expected musclememapi.Exercise
		err      error
	}{
		{"ErrorOnEmptyJSON", []byte("{}"), empty, musclememapi.ErrEmptyExercise},
		{"ErrorOnInvalidJSON", []byte("{\"invalid\":true"), empty, musclememapi.ErrInvalidJSON},
		{"ErrorOnNotAnExercise", []byte("{\"unknown\":1}"), empty, musclememapi.ErrEmptyExercise},
		{"ErrorOnNonExistingId", e3JSON, empty, musclememapi.ErrExerciseNotFound},
		{"CreateNewExerciseOnNoID", e1JSON, e1, nil},
		{"OverwriteExerciseOnExistingId", e2JSON, e2, nil},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {
			err := musclememapi.StoreExerciseJSON(xs, c.given)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestLinkedListOfExercises(t *testing.T) {
	xs := newTestSqliteDatastore(t)
	defer xs.Close()

	xsJSON, err := json.Marshal(xs)
	if err != nil {
		t.Error(err)
	}

	cs := []struct {
		testname string
		repo     musclememapi.ExerciseStorer
		given    []byte
		expected any
		err      error
	}{
		{"InsertAlwaysAtTail", xs, xsJSON, nil, nil},
	}

	for _, c := range cs {
		t.Run(c.testname, func(t *testing.T) {

		})
	}
}
