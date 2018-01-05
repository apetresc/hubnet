package main

import (
	"log"
	"net"
	"sync"

	"github.com/dustin/go-nntp"
	"github.com/dustin/go-nntp/server"
)

func maybefatal(err error, f string, a ...interface{}) {
	if err != nil {
		log.Fatalf(f, a...)
	}
}

type githubBackend struct {
	groups    map[string]*nntp.Group
	grouplock sync.Mutex
}

func (gb *githubBackend) AllowPost() bool {
	return true
}

func (gb *githubBackend) Authenticate(user, pass string) (nntpserver.Backend, error) {
	return nil, nntpserver.ErrAuthRejected
}

func (gb *githubBackend) Authorized() bool {
	return true
}

func (gb *githubBackend) GetArticle(group *nntp.Group, id string) (*nntp.Article, error) {
	return nil, nil
}

func (gb *githubBackend) GetArticles(group *nntp.Group, from, to int64) ([]nntpserver.NumberedArticle, error) {
	return nil, nil
}

func (gb *githubBackend) GetGroup(name string) (*nntp.Group, error) {
	return nil, nil
}

func (gb *githubBackend) ListGroups(max int) ([]*nntp.Group, error) {
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

func (gb *githubBackend) Post(art *nntp.Article) error {
	return nil
}

func main() {
	backend := githubBackend{}

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
