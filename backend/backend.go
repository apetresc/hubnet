package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

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
	var numberedArticles []nntpserver.NumberedArticle
	var i int
	var err error
	if numberedArticles, err = sb.GetStoredArticles(group, 0, 9999); err != nil {
		log.Fatal(err)
	} else {
		i, _ = strconv.Atoi(id)
	}

	return numberedArticles[i-1].Article, nil
}

func (sb *SQLBackend) GetStoredArticles(group *nntp.Group, from, to int64) ([]nntpserver.NumberedArticle, error) {
	rv := make([]nntpserver.NumberedArticle, 0, 0)
	rows, err := sb.DB.Query(fmt.Sprintf(`
SELECT messageid, subject, body, author, date, refs
FROM articles a JOIN newsgroups g ON a.newsgroup = g.id AND a.type = g.type
WHERE g.name = "%s"`,
		group.Name))
	defer rows.Close()

	if err != nil {
		log.Fatalf("Error listing articles: %v", err)
	}
	id := int64(1)
	for rows.Next() {
		var messageid, subject, body, author, refs string
		var date int64
		if err := rows.Scan(&messageid, &subject, &body, &author, &date, &refs); err != nil {
			log.Fatal(err)
		}
		headers := make(textproto.MIMEHeader)
		headers.Add("Message-Id", fmt.Sprintf("<%s>", messageid))
		headers.Add("Date", time.Unix(date, 0).Format(time.RFC850))
		headers.Add("From", author)
		headers.Add("Subject", subject)
		headers.Add("Newsgroups", group.Name)
		rv = append(rv, nntpserver.NumberedArticle{
			Num: id,
			Article: &nntp.Article{
				Header: headers,
				Body:   strings.NewReader(body),
				Bytes:  len(body),
				Lines:  strings.Count(body, "\n") + 1,
			},
		})
		id += 1
		if id > to {
			break
		}
	}
	return rv, nil
}

func (sb *SQLBackend) GetArticles(group *nntp.Group, from, to int64) ([]nntpserver.NumberedArticle, error) {
	// First let's do a fetch
	var repoName = strings.Join(strings.Split(group.Name, ".")[2:], "/")
	fetchRepo(sb, repoName)

	return sb.GetStoredArticles(group, from, to)
}

func (sb *SQLBackend) GetGroup(name string) (*nntp.Group, error) {
	row := sb.DB.QueryRow("SELECT name, type FROM newsgroups WHERE name=\"" + name + "\"")
	var _type string
	if err := row.Scan(&name, &_type); err != nil {
		log.Fatal(err)
	}

	var group = &nntp.Group{
		Name:        name,
		Description: fmt.Sprintf("%ss for the Github repository at ", _type),
		Count:       0,
		Low:         0,
		High:        0,
		Posting:     nntp.PostingNotPermitted,
	}

	var articles, err = sb.GetStoredArticles(group, 0, 9999)
	if err != nil {
		return nil, err
	}
	if len(articles) > 0 {
		group.Low = articles[0].Num
		group.Count = int64(len(articles))
		group.High = articles[len(articles)-1].Num
	}

	fmt.Println("RETURNING GROUP:", group)

	return group, nil

}

func (sb *SQLBackend) ListGroups(max int) ([]*nntp.Group, error) {
	// First we fetch
	fetchAllGroups(sb)

	rv := make([]*nntp.Group, 0, 0)
	rows, err := sb.DB.Query("SELECT name, type FROM newsgroups")
	defer rows.Close()
	if err != nil {
		log.Fatalf("Error listing newsgroups: %v", err)
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
