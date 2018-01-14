package main

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
)

func main() {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 1, 99, "")

	if err != nil {

	}

	github.NewClient(&http.Client{Transport: itr})
}
