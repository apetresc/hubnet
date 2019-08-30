package main

import (
	"context"
	"database/sql"
	"flag"
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

type Repository struct {
	Id               string
	NameWithOwner    string
	HasIssuesEnabled bool
}

type PageInfo struct {
	EndCursor   string
	HasNextPage bool
}

type Author struct {
	Login string
}

type Issue struct {
	Id       string
	Author   Author
	Title    string
	BodyText string
}

type PullRequest struct {
	Id       string
	Author   Author
	Title    string
	BodyText string
	Comments struct {
		Nodes    []Comment
		PageInfo PageInfo
	} `graphql:"comments(first:100, after:$pullRequestCommentsCursor)"`
}

type Comment struct {
	BodyText string
	Author   Author
}

func addRepository(sb *backend.SQLBackend, repo Repository) error {
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

func addIssueArticle(sb *backend.SQLBackend, issue Issue) error {
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO articles(messageid, author, subject, date, refs) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(issue.Id, issue.Author.Login, issue.Title, "", "")
	if err != nil {
		if sqlerr, ok := err.(sqlite3.Error); ok && sqlerr.ExtendedCode == 1555 {
			log.Printf("Skipping over %s, already exists...\n", issue.Id)
		} else {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func fetchAllGroups(sb *backend.SQLBackend) error {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	client := githubv4.NewClient(oauth2.NewClient(context.Background(), src))

	var q struct {
		Viewer struct {
			Login        string
			CreatedAt    time.Time
			Repositories struct {
				Nodes    []Repository
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
	var allRepos []Repository
	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			// Handle error
			fmt.Println(err)
		}
		for _, repo := range q.Viewer.Repositories.Nodes {
			fmt.Println("        Repo:", repo.NameWithOwner)
			fmt.Println("        Issues:", repo.HasIssuesEnabled)
			addRepository(sb, repo)
		}
		allRepos = append(allRepos, q.Viewer.Repositories.Nodes...)
		if !q.Viewer.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Viewer.Repositories.PageInfo.EndCursor)
	}
	fmt.Println("Total # of repos:", len(allRepos))

	return nil
}

func fetchRepo(sb *backend.SQLBackend, repoName string) error {
	var strs = strings.SplitN(repoName, "/", 2)
	var owner = strs[0]
	var name = strs[1]
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	client := githubv4.NewClient(oauth2.NewClient(context.Background(), src))

	var q struct {
		Repository struct {
			Id     string
			Issues struct {
				Nodes    []Issue
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"issues(first:100, after:$issuesCursor)"`
			PullRequests struct {
				Nodes    []PullRequest
				PageInfo PageInfo
			} `graphql:"pullRequests(first:100, after:$pullRequestsCursor)"`
		} `graphql:"repository(owner:$owner, name:$name)"`
	}
	variables := map[string]interface{}{
		"owner":                     githubv4.String(owner),
		"name":                      githubv4.String(name),
		"issuesCursor":              (*githubv4.String)(nil),
		"pullRequestsCursor":        (*githubv4.String)(nil),
		"pullRequestCommentsCursor": (*githubv4.String)(nil),
	}

	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			fmt.Println("Errorrr: ", err)
			return err
		}

		fmt.Println("Repo:", q.Repository.Id)
		var allIssues []Issue
		var allPRs []PullRequest
		for _, issue := range q.Repository.Issues.Nodes {
			fmt.Printf("Issue(%s): %s\n", issue.Author.Login, issue.Title)
			if err = addIssueArticle(sb, issue); err != nil {
				log.Fatal(err)
				return err
			}
		}
		allIssues = append(allIssues, q.Repository.Issues.Nodes...)
		for _, pullRequest := range q.Repository.PullRequests.Nodes {
			fmt.Printf("PR(%s): %s\n", pullRequest.Author.Login, pullRequest.Title)
			for _, pullRequestComment := range pullRequest.Comments.Nodes {
				fmt.Printf("\tComment(%s): %s\n", pullRequestComment.Author.Login, pullRequestComment.BodyText)
			}
		}
		allPRs = append(allPRs, q.Repository.PullRequests.Nodes...)

		break
	}

	return nil
}

func main() {
	var repo = flag.String("repo", "", "repo to fetch")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./hubnet.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	backend.EnsureViews(db)
	backend := backend.SQLBackend{
		DB: db,
	}

	if *repo == "" {
		if err = fetchAllGroups(&backend); err != nil {
			log.Fatal(err)
		}
	} else {
		fetchRepo(&backend, *repo)
	}
}
