package backend

import (
	"database/sql"
	"errors"
	"log"
)

const createGroupsQuery = `
  CREATE TABLE IF NOT EXISTS newsgroups (
		id text not null,
		name text,
		type text,
		description text,
		primary key (id, type)
	);
`

const createArticlesQuery = `
  CREATE TABLE IF NOT EXISTS articles (
		messageid text,
		subject text,
		body text,
		author text,
		date integer,
		refs text,
		newsgroup text not null,
		foreign key (newsgroup) references newsgroups(id),
		primary key (messageid)
  );
`

func EnsureViews(db *sql.DB) error {
	_, errg := db.Exec(createGroupsQuery)
	if errg != nil {
		log.Printf("Error creating newsgroups view: %v\n", errg)
		return errors.New("Error creating views")
	}
	_, errg = db.Exec(createArticlesQuery)
	if errg != nil {
		log.Printf("Error creating articles view: %v\n", errg)
		return errors.New("Error creating views")
	}

	return nil
}
