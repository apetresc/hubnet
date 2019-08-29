package backend

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	nntp "github.com/dustin/go-nntp"
	"github.com/dustin/go-nntp/server"
)

type SQLBackend struct {
	DB        *sql.DB
	Groups    map[string]*nntp.Group
	Grouplock sync.Mutex
}

func (sb *SQLBackend) AllowPost() bool {
	return false
}

func (sb *SQLBackend) Authenticate(user, pass string) (nntpserver.Backend, error) {
	return nil, nntpserver.ErrAuthRejected
}

func (sb *SQLBackend) Authorized() bool {
	return true
}

func (sb *SQLBackend) GetArticle(group *nntp.Group, id string) (*nntp.Article, error) {
	return nil, nil
}

func (sb *SQLBackend) GetArticles(group *nntp.Group, from, to int64) ([]nntpserver.NumberedArticle, error) {
	rv := make([]nntpserver.NumberedArticle, 0, 0)
	//rv = append(rv, nntpserver.NumberedArticle{})
	return rv, nil
}

func (sb *SQLBackend) GetGroup(name string) (*nntp.Group, error) {
	row := sb.DB.QueryRow("SELECT name, type FROM groups WHERE name=\"" + name + "\"")
	var _type string
	if err := row.Scan(&name, &_type); err != nil {
		log.Fatal(err)
	}
	return &nntp.Group{
		Name:        name,
		Description: fmt.Sprintf("%ss for the Github repository at ", _type),
		Count:       0,
		Low:         0,
		High:        0,
		Posting:     nntp.PostingNotPermitted,
	}, nil
}

func (sb *SQLBackend) ListGroups(max int) ([]*nntp.Group, error) {
	rv := make([]*nntp.Group, 0, 0)
	rows, err := sb.DB.Query("SELECT name, type FROM groups")
	defer rows.Close()
	if err != nil {
		log.Fatalf("Error listing groups: %v", err)
	}
	for rows.Next() {
		var name string
		var _type string
		if err := rows.Scan(&name, &_type); err != nil {
			log.Fatal(err)
		}
		rv = append(rv, &nntp.Group{
			Name:        name,
			Description: fmt.Sprintf("%ss for the Github repository at ", _type),
			Count:       0,
			Low:         0,
			High:        0,
			Posting:     nntp.PostingNotPermitted,
		})
	}
	return rv, nil
}

func (sb *SQLBackend) Post(art *nntp.Article) error {
	return nil
}
