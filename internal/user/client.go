package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/scrot/musclemem-api/internal/sdk"
)

type UserClient struct {
	*sdk.Client
}

func NewUserClient(client *sdk.Client) *UserClient {
	return &UserClient{Client: client}
}

func (c *UserClient) Login() {
}

func (c *UserClient) Register(ctx context.Context, u User) (User, *http.Response, error) {
	path := "/users"

	userJSON, err := json.Marshal(u)
	if err != nil {
		return User{}, nil, err
	}

	resp, err := c.Send(ctx, http.MethodPost, path, bytes.NewReader(userJSON))
	if err != nil {
		return User{}, nil, err
	}

	var respUser User
	if err := json.NewDecoder(resp.Body).Decode(&respUser); err != nil {
		return User{}, nil, err
	}

	return respUser, resp, nil
}
