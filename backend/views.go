package backend

import (
	"database/sql"
	"errors"
	"log"
)

const create_groups_query = `
  CREATE TABLE IF NOT EXISTS groups (
		id integer not null primary key,
		name text,
		description text)
	;
`

func EnsureViews(db *sql.DB) error {
	_, errg := db.Exec(create_groups_query)
	if errg != nil {
		log.Printf("Error creating groups view: %v\n", errg)
	}

	if errg != nil {
		return errors.New("Error creating views")
	}

	return nil
}
