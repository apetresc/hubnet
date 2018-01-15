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
	_ "github.com/mattn/go-sqlite3"
)

func addRepository(sb *backend.SQLBackend, repo *github.Repository) error {
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO groups(id, name) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, groupType := range [2]string{"prs", "issues"} {
		_, err = stmt.Exec(
			repo.GetID(),
			fmt.Sprintf("github.%s.%s.%s", groupType, repo.GetOwner().GetLogin(), repo.GetName()))
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func main() {
	args := os.Args[1:]
	keyPath, _ := filepath.Abs(args[0])
	fmt.Printf("keyPath is %s\n", keyPath)

	db, err := sql.Open("sqlite3", "./usehub.db")
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
			addRepository(&backend, repo)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

}