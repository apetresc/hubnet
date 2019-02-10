package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/apetresc/usehub/backend"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/mattn/go-sqlite3"
	"launchpad.net/go-xdg"
)

func addRepository(sb *backend.SQLBackend, repo *github.Repository) error {
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO groups(id, type, name) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, groupType := range [2]string{"prs", "issues"} {
		_, err = stmt.Exec(
			repo.GetID(),
			groupType,
			fmt.Sprintf("github.%s.%s.%s", groupType, repo.GetOwner().GetLogin(), repo.GetName()))
		if err != nil {
			if sqlerr, ok := err.(sqlite3.Error); ok && sqlerr.ExtendedCode == 1555 {
				log.Printf("Skipping over %s, already exists...\n", repo.GetFullName())
			} else {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	args := os.Args[1:]
	keyPath, _ := filepath.Abs(args[0])

	db, err := sql.Open("sqlite3", filepath.Join(xdg.Data.Home(), "usehub", "usehub.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	backend.EnsureViews(db)
	backend := backend.SQLBackend{
		DB: db,
	}

	ctx := context.Background()
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 7898, 80268, keyPath)

	if err != nil {
		log.Fatal(err)
	}

	client := github.NewClient(&http.Client{Transport: itr})

	opt := &github.RepositoryListOptions{
		Visibility:  "all",
		Affiliation: "owner,collaborator,organization_member",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.List(ctx, "apetresc", opt)
		if err != nil {
			log.Fatal(err)
		}
		for _, repo := range repos {
			fmt.Printf("%s\n", repo)
			err = addRepository(&backend, repo)
			if err != nil {
				log.Fatal(err)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

}
