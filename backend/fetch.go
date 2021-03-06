package backend

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	Id        string
	Author    Author
	CreatedAt time.Time
	Title     string
	Body      string
	Comments  struct {
		Nodes    []Comment
		PageInfo PageInfo
	} `graphql:"comments(first:100, after:$issuesCommentsCursor)"`
}

type PullRequest struct {
	Id        string
	Author    Author
	CreatedAt time.Time
	Title     string
	Body      string
	Comments  struct {
		Nodes    []Comment
		PageInfo PageInfo
	} `graphql:"comments(first:100, after:$pullRequestsCommentsCursor)"`
}

type Comment struct {
	Id        string
	Author    Author
	CreatedAt time.Time
	Body      string
}

func addRepository(sb *SQLBackend, repo Repository) error {
	strs := strings.SplitN(repo.NameWithOwner, "/", 2)
	owner := strs[0]
	name := strs[1]
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		"INSERT INTO newsgroups(id, type, name) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, groupType := range [2]string{"pr", "issue"} {
		_, err = stmt.Exec(
			repo.Id,
			groupType,
			fmt.Sprintf("github.%s.%s.%s", groupType, owner, name))
		if err != nil {
			if sqlerr, ok := err.(sqlite3.Error); ok &&
				sqlerr.ExtendedCode == 1555 {
				log.Printf("Skipping over %s, already exists...\n",
					repo.NameWithOwner)
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

func addIssueArticle(sb *SQLBackend, issue Issue, repository string) error {
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO articles(messageid,
							 author,
							 subject,
							 body,
							 date,
							 refs,
							 newsgroup,
							 type)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(messageid) DO UPDATE SET body=excluded.body`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		issue.Id,
		issue.Author.Login,
		issue.Title,
		issue.Body,
		issue.CreatedAt.Unix(),
		"",
		repository,
		"issue",
	)
	if err != nil {
		if sqlerr, ok := err.(sqlite3.Error); ok &&
			sqlerr.ExtendedCode == 1555 {
			log.Printf("Skipping over %s, already exists...\n", issue.Id)
		} else {
			return err
		}
	}

	for _, issueComment := range issue.Comments.Nodes {
		_, err = stmt.Exec(
			issueComment.Id,
			issueComment.Author.Login,
			fmt.Sprintf("Re: %s", issue.Title),
			issueComment.Body,
			issueComment.CreatedAt.Unix(),
			issue.Id,
			repository,
			"issue",
		)
		if err != nil {
			if sqlerr, ok := err.(sqlite3.Error); ok &&
				sqlerr.ExtendedCode == 1555 {
				log.Printf("Skipping over %s, already exists...\n", issue.Id)
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

func addPRArticle(sb *SQLBackend, pr PullRequest, repository string) error {
	tx, err := sb.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO articles(messageid,
		author,
		subject,
		body,
		date,
		refs,
		newsgroup,
		type)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(messageid) DO UPDATE SET body=excluded.body`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		pr.Id,
		pr.Author.Login,
		pr.Title,
		pr.Body,
		pr.CreatedAt.Unix(),
		"",
		repository,
		"pr",
	)
	if err != nil {
		if sqlerr, ok := err.(sqlite3.Error); ok &&
			sqlerr.ExtendedCode == 1555 {
			log.Printf("Skipping over %s, already exists...\n", pr.Id)
		} else {
			return err
		}
	}

	for _, prComment := range pr.Comments.Nodes {
		_, err = stmt.Exec(
			prComment.Id,
			prComment.Author.Login,
			fmt.Sprintf("Re: %s", pr.Title),
			prComment.Body,
			prComment.CreatedAt.Unix(),
			pr.Id,
			repository,
			"pr",
		)
		if err != nil {
			if sqlerr, ok := err.(sqlite3.Error); ok &&
				sqlerr.ExtendedCode == 1555 {
				log.Printf("Skipping over %s, already exists...\n", pr.Id)
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

func fetchAllGroups(sb *SQLBackend) error {
	src := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: os.Getenv("GITHUB_TOKEN")})
	client := githubv4.NewClient(oauth2.NewClient(context.Background(), src))

	var q struct {
		Viewer struct {
			Login        string
			CreatedAt    time.Time
			Repositories struct {
				Nodes    []Repository
				PageInfo PageInfo
			} `graphql:"repositories(first:100, after:$reposCursor)"`
			StarredRepositories struct {
				Nodes    []Repository
				PageInfo PageInfo
			} `graphql:"starredRepositories(first:100, after:$starredCursor)"`
		}
	}
	variables := map[string]interface{}{
		"reposCursor":   (*githubv4.String)(nil),
		"starredCursor": (*githubv4.String)(nil),
	}
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
		for _, repo := range q.Viewer.StarredRepositories.Nodes {
			fmt.Println("        Repo:", repo.NameWithOwner)
			fmt.Println("        Issues:", repo.HasIssuesEnabled)
			addRepository(sb, repo)
		}
		if !q.Viewer.Repositories.PageInfo.HasNextPage &&
			!q.Viewer.StarredRepositories.PageInfo.HasNextPage {
			break
		}
		if len(q.Viewer.Repositories.PageInfo.EndCursor) > 0 {
			variables["reposCursor"] = githubv4.String(
				q.Viewer.Repositories.PageInfo.EndCursor)
		}
		if len(q.Viewer.StarredRepositories.PageInfo.EndCursor) > 0 {
			variables["starredCursor"] = githubv4.String(
				q.Viewer.StarredRepositories.PageInfo.EndCursor)
		}
	}

	return nil
}

func fetchRepo(sb *SQLBackend, repoName string) error {
	var strs = strings.SplitN(repoName, "/", 2)
	var owner = strs[0]
	var name = strs[1]
	src := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: os.Getenv("GITHUB_TOKEN")})
	client := githubv4.NewClient(oauth2.NewClient(context.Background(), src))

	var q struct {
		Repository struct {
			Id     string
			Issues struct {
				Nodes    []Issue
				PageInfo PageInfo
			} `graphql:"issues(first:100, after:$issuesCursor)"`
			PullRequests struct {
				Nodes    []PullRequest
				PageInfo PageInfo
			} `graphql:"pullRequests(first:100, after:$pullRequestsCursor)"`
		} `graphql:"repository(owner:$owner, name:$name)"`
	}
	variables := map[string]interface{}{
		"owner":                      githubv4.String(owner),
		"name":                       githubv4.String(name),
		"issuesCursor":               (*githubv4.String)(nil),
		"issuesCommentsCursor":       (*githubv4.String)(nil),
		"pullRequestsCursor":         (*githubv4.String)(nil),
		"pullRequestsCommentsCursor": (*githubv4.String)(nil),
	}

	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			fmt.Println("Errorrr: ", err)
			return err
		}

		fmt.Println("Repo:", q.Repository.Id)
		for _, issue := range q.Repository.Issues.Nodes {
			fmt.Printf("Issue(%s): %s\n", issue.Author.Login, issue.Title)
			if err = addIssueArticle(sb, issue, q.Repository.Id); err != nil {
				log.Fatal(err)
				return err
			}
			for _, issueComment := range issue.Comments.Nodes {
				fmt.Printf("\tComment(%s): %s\n",
					issueComment.Author.Login, issueComment.Body)
			}
		}
		for _, pullRequest := range q.Repository.PullRequests.Nodes {
			fmt.Printf("PR(%s): %s\n",
				pullRequest.Author.Login, pullRequest.Title)
			if err = addPRArticle(
				sb,
				pullRequest,
				q.Repository.Id); err != nil {
				log.Fatal(err)
				return err
			}
			for _, pullRequestComment := range pullRequest.Comments.Nodes {
				fmt.Printf("\tComment(%s): %s\n",
					pullRequestComment.Author.Login, pullRequestComment.Body)
			}
		}

		if !q.Repository.Issues.PageInfo.HasNextPage &&
			!q.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		if len(q.Repository.Issues.PageInfo.EndCursor) > 0 {
			variables["issuesCursor"] = githubv4.String(
				q.Repository.Issues.PageInfo.EndCursor)
		}
		if len(q.Repository.PullRequests.PageInfo.EndCursor) > 0 {
			variables["pullRequestsCursor"] = githubv4.String(
				q.Repository.PullRequests.PageInfo.EndCursor)
		}
	}
	return nil
}

func oldmain() {
	var repo = flag.String("repo", "", "repo to fetch")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./hubnet.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	EnsureViews(db)
	backend := SQLBackend{
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
