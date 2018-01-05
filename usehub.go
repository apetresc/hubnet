package main

import (
	"database/sql"
	"log"
	"net"
	"sync"

	"github.com/dustin/go-nntp"
	"github.com/dustin/go-nntp/server"
	_ "github.com/mattn/go-sqlite3"
)

func maybefatal(err error, f string, a ...interface{}) {
	if err != nil {
		log.Fatalf(f, a...)
	}
}

type SQLBackend struct {
	groups    map[string]*nntp.Group
	grouplock sync.Mutex
}

func (sb *SQLBackend) AllowPost() bool {
	return true
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
	return nil, nil
}

func (sb *SQLBackend) GetGroup(name string) (*nntp.Group, error) {
	return nil, nil
}

func (sb *SQLBackend) ListGroups(max int) ([]*nntp.Group, error) {
	rv := make([]*nntp.Group, 0, 0)
	rv = append(rv, &nntp.Group{
		Name:        "rec.games.awesome",
		Description: "A cool test server",
		Count:       2,
		Low:         2,
		High:        3,
		Posting:     nntp.PostingPermitted,
	})
	return rv, nil
}

func (sb *SQLBackend) Post(art *nntp.Article) error {
	return nil
}

func main() {
	backend := SQLBackend{}

	a, err := net.ResolveTCPAddr("tcp", ":1119")
	maybefatal(err, "Error resolving listener: %v", err)
	l, err := net.ListenTCP("tcp", a)
	maybefatal(err, "Error setting up listener: %v", err)
	defer l.Close()

	s := nntpserver.NewServer(&backend)

	for {
		c, err := l.AcceptTCP()
		maybefatal(err, "Error accepting connection: %v", err)
		go s.Process(c)
	}
}
