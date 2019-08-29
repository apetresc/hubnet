package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/apetresc/hubnet/backend"
	"github.com/mattn/go-sqlite3"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type repository struct {
	Id               string
	NameWithOwner    string
	HasIssuesEnabled bool
}

func addRepository(sb *backend.SQLBackend, repo repository) error {
	strs := strings.SplitN(repo.NameWithOwner, "/", 2)
	owner := strs[0]
	name := strs[1]
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
			repo.Id,
			groupType,
			fmt.Sprintf("github.%s.%s.%s", groupType, owner, name))
		if err != nil {
			if sqlerr, ok := err.(sqlite3.Error); ok && sqlerr.ExtendedCode == 1555 {
				log.Printf("Skipping over %s, already exists...\n", repo.NameWithOwner)
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
	db, err := sql.Open("sqlite3", "./hubnet.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	backend.EnsureViews(db)
	backend := backend.SQLBackend{
		DB: db,
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	client := githubv4.NewClient(oauth2.NewClient(context.Background(), src))

	var q struct {
		Viewer struct {
			Login        string
			CreatedAt    time.Time
			Repositories struct {
				Nodes    []repository
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"repositories(first:100, after:$commentsCursor)"`
		}
	}
	variables := map[string]interface{}{
		"commentsCursor": (*githubv4.String)(nil),
	}
	var allRepos []repository
	for {
		err = client.Query(context.Background(), &q, variables)
		if err != nil {
			// Handle error
			fmt.Println(err)
		}
		for _, repo := range q.Viewer.Repositories.Nodes {
			fmt.Println("        Repo:", repo.NameWithOwner)
			fmt.Println("        Issues:", repo.HasIssuesEnabled)
			addRepository(&backend, repo)
		}
		allRepos = append(allRepos, q.Viewer.Repositories.Nodes...)
		if !q.Viewer.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Viewer.Repositories.PageInfo.EndCursor)
	}
	fmt.Println("Total # of repos:", len(allRepos))
}
