package workout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scrot/musclemem-api/internal/sdk"
)

type WorkoutClient struct {
	*sdk.Client
}

func NewWorkoutClient(client *sdk.Client) *WorkoutClient {
	return &WorkoutClient{Client: client}
}

func (c *WorkoutClient) List(ctx context.Context, username string) ([]Workout, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts", username)

	resp, err := c.Send(ctx, http.MethodGet, path, nil)
	if err != nil {
		return []Workout{}, resp, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	var ws []Workout
	if err := dec.Decode(&ws); err != nil {
		return []Workout{}, resp, err
	}

	return ws, resp, nil
}

func (c *WorkoutClient) Add(ctx context.Context, w Workout) (Workout, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts", w.Owner)

	body, err := json.Marshal(w)
	if err != nil {
		return Workout{}, nil, err
	}

	resp, err := c.Send(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return Workout{}, resp, err
	}

	var respWorkout Workout
	if err := json.NewDecoder(resp.Body).Decode(&respWorkout); err != nil {
		return Workout{}, resp, err
	}

	return respWorkout, resp, nil
}

func (c *WorkoutClient) Delete(ctx context.Context, w WorkoutRef) (Workout, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d", w.Username, w.WorkoutIndex)

	resp, err := c.Send(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return Workout{}, nil, err
	}

	var respWorkout Workout
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&respWorkout); err != nil {
		return Workout{}, nil, err
	}

	return respWorkout, resp, nil
}

func (c *WorkoutClient) Update(ctx context.Context, ref WorkoutRef, w Workout) (Workout, *http.Response, error) {
	path := fmt.Sprintf("/users/%s/workouts/%d", ref.Username, ref.WorkoutIndex)

	workoutJSON, err := json.Marshal(w)
	if err != nil {
		return Workout{}, nil, err
	}

	resp, err := c.Send(ctx, http.MethodPatch, path, bytes.NewReader(workoutJSON))
	if err != nil {
		return Workout{}, nil, err
	}

	var respWorkout Workout
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&respWorkout); err != nil {
		return Workout{}, nil, err
	}

	return respWorkout, resp, nil
}
