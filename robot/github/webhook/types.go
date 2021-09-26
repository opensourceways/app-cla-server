package webhook

import sdk "github.com/google/go-github/v36/github"

type GenericEvent struct {
	Sender sdk.User       `json:"sender"`
	Repo   sdk.Repository `json:"repository"`
}
