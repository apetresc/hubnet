package backend

import (
	"database/sql"
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
	rv = append(rv, nntpserver.NumberedArticle{})
	return rv, nil
}

func (sb *SQLBackend) GetGroup(name string) (*nntp.Group, error) {
	return &nntp.Group{
		Name:        "rec.games.awesome",
		Description: "A cool test server",
		Count:       2,
		Low:         2,
		High:        3,
		Posting:     nntp.PostingNotPermitted,
	}, nil
}

func (sb *SQLBackend) ListGroups(max int) ([]*nntp.Group, error) {
	rv := make([]*nntp.Group, 0, 0)
	rv = append(rv, &nntp.Group{
		Name:        "rec.games.awesome",
		Description: "A cool test server",
		Count:       2,
		Low:         2,
		High:        3,
		Posting:     nntp.PostingNotPermitted,
	})
	return rv, nil
}

func (sb *SQLBackend) Post(art *nntp.Article) error {
	return nil
}
