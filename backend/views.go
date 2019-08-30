package backend

import (
	"database/sql"
	"errors"
	"log"
)

const createGroupsQuery = `
  CREATE TABLE IF NOT EXISTS groups (
		id text not null,
		name text,
		type text,
		description text,
		primary key (id, type)
	);
`

const createArticlesQuery = `
  CREATE TABLE IF NOT EXISTS articles (
		id integer not null,
		messageid text,
		subject text,
		author text,
		date integer,
		refs text,
		primary key (id)
  );
`

func EnsureViews(db *sql.DB) error {
	_, errg := db.Exec(createGroupsQuery)
	if errg != nil {
		log.Printf("Error creating groups view: %v\n", errg)
		return errors.New("Error creating views")
	}
	_, errg = db.Exec(createArticlesQuery)
	if errg != nil {
		log.Printf("Error creating articles view: %v\n", errg)
		return errors.New("Error creating views")
	}

	return nil
}
