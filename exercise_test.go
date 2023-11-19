package musclememapi_test

// import (
// 	"encoding/json"
// 	"testing"

// 	musclememapi "github.com/scrot/musclemem-api"
// 	"github.com/stretchr/testify/assert"
// )

// func newTestSqliteDatastore(t *testing.T, xs ...musclememapi.Exercise) *musclememapi.SqliteDatastore {
// 	t.Helper()

// 	cfg := musclememapi.SqliteDatastoreConfig{"file://" + t.TempDir(), true, nil}
// 	db, err := musclememapi.NewSqliteDatastore(cfg)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	return db
// }

// func TestFetchSingleExerciseJSON(t *testing.T) {
// 	e1 := musclememapi.Exercise{1, "Benchpress", 90.5, 6, nil, nil}
// 	e1JSON, _ := json.Marshal(&e1)

// 	var xs musclememapi.Datastore

// 	cs := []struct {
// 		testname string
// 		given    int
// 		expected []byte
// 		err      error
// 	}{
// 		{"ErrorOnNoId", 0, musclememapi.EmptyJSON, musclememapi.ErrNoIDProvided},
// 		{"ErrorOnInvalidIdFormat", -1, musclememapi.EmptyJSON, musclememapi.ErrInvalidIdFormat},
// 		{"ErrorOnIdNotFound", 2, musclememapi.EmptyJSON, musclememapi.ErrExerciseNotFound},
// 		{"EqExerciseOnExistingID", e1.ID, e1JSON, nil},
// 	}

// 	for _, c := range cs {
// 		t.Run(c.testname, func(t *testing.T) {
// 			actual, err := musclememapi.FetchSingleExerciseJSON(&xs, c.given)
// 			if c.err != nil {
// 				assert.ErrorIs(t, err, c.err)
// 			}
// 			assert.Equal(t, c.expected, actual)
// 		})
// 	}
// }

// func TestStoreSingleExerciseJSON(t *testing.T) {
// 	// empty repository
// 	xs := newTestSqliteDatastore(t)
// 	defer xs.Close()

// 	// test exercises
// 	empty := musclememapi.Exercise{}

// 	e1 := musclememapi.Exercise{0, "New", 90.5, 6, nil, nil}
// 	e1JSON, _ := json.Marshal(&e1)

// 	e2 := musclememapi.Exercise{1, "Overwritten", 90.5, 6, nil, nil}
// 	e2JSON, _ := json.Marshal(&e2)

// 	e3 := musclememapi.Exercise{2, "Error", 90.5, 6, nil, nil}
// 	e3JSON, _ := json.Marshal(&e3)

// 	// insert e2

// 	cs := []struct {
// 		testname string
// 		given    []byte
// 		expected musclememapi.Exercise
// 		err      error
// 	}{
// 		{"ErrorOnEmptyJSON", []byte("{}"), empty, musclememapi.ErrEmptyExercise},
// 		{"ErrorOnInvalidJSON", []byte("{\"invalid\":true"), empty, musclememapi.ErrInvalidJSON},
// 		{"ErrorOnNotAnExercise", []byte("{\"unknown\":1}"), empty, musclememapi.ErrEmptyExercise},
// 		{"ErrorOnNonExistingId", e3JSON, empty, musclememapi.ErrExerciseNotFound},
// 		{"CreateNewExerciseOnNoID", e1JSON, e1, nil},
// 		{"OverwriteExerciseOnExistingId", e2JSON, e2, nil},
// 	}

// 	for _, c := range cs {
// 		t.Run(c.testname, func(t *testing.T) {
// 			err := musclememapi.StoreExerciseJSON(xs, c.given)
// 			assert.ErrorIs(t, err, c.err)
// 		})
