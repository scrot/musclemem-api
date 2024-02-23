package exercise

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/scrot/musclemem-api/internal/sdk"
	"github.com/scrot/musclemem-api/internal/workout"
)

type ExerciseClient struct {
	*sdk.Client
}

func NewExerciseClient(client *sdk.Client) *ExerciseClient {
	return &ExerciseClient{Client: client}
}

func (c *ExerciseClient) List(ctx context.Context, ref workout.WorkoutRef) ([]Exercise, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d/exercises", ref.Username, ref.WorkoutIndex)

	resp, err := c.Send(ctx, http.MethodGet, path, nil)
	if err != nil {
		return []Exercise{}, resp, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var xs []Exercise
	if err := dec.Decode(&xs); err != nil {
		return []Exercise{}, resp, err
	}

	return xs, resp, nil
}

func (c *ExerciseClient) Add(ctx context.Context, e Exercise) (Exercise, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d/exercises", e.Owner, e.Workout)

	body, err := json.Marshal(e)
	if err != nil {
		return Exercise{}, nil, err
	}

	resp, err := c.Send(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return Exercise{}, resp, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var x Exercise
	if err := dec.Decode(&x); err != nil {
		return Exercise{}, resp, err
	}

	return x, resp, nil
}

func (c *ExerciseClient) Delete(ctx context.Context, ref ExerciseRef) (*http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", ref.Username, ref.WorkoutIndex, ref.ExerciseIndex)

	resp, err := c.Send(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return resp, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var x Exercise
	if err := dec.Decode(&x); err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *ExerciseClient) Update(ctx context.Context, ref ExerciseRef, patch Exercise) (Exercise, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", ref.Username, ref.WorkoutIndex, ref.ExerciseIndex)

	body, err := json.Marshal(patch)
	if err != nil {
		return Exercise{}, nil, err
	}

	resp, err := c.Send(ctx, http.MethodPatch, path, bytes.NewReader(body))
	if err != nil {
		return Exercise{}, resp, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var x Exercise
	if err := dec.Decode(&x); err != nil {
		return Exercise{}, resp, err
	}

	return x, resp, nil
}

type Move string

const (
	MoveDown Move = "down"
	MoveUp   Move = "up"
	MoveSwap Move = "swap"
)

func (c *ExerciseClient) Move(ctx context.Context, ref ExerciseRef, dir Move, with *ExerciseRef) (*http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/%s", ref.Username, ref.WorkoutIndex, ref.ExerciseIndex, dir)

	if dir == MoveSwap {
		if with == nil {
			return nil, errors.New("swap requires with reference")
		}
	} else {
		if with != nil {
			return nil, errors.New(string(dir) + " requires no with reference")
		}
	}

	resp, err := c.Send(ctx, http.MethodPut, path, strings.NewReader((string(dir))))
	if err != nil {
		return resp, err
	}

	return resp, nil
}
