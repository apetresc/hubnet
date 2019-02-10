package backend

import (
	"database/sql"
	"errors"
	"log"
)

const createGroupsQuery = `
  CREATE TABLE IF NOT EXISTS groups (
		id integer not null,
		name text,
		type text,
		description text,
		primary key (id, type)
	);
`

func EnsureViews(db *sql.DB) error {
	_, errg := db.Exec(createGroupsQuery)
	if errg != nil {
		log.Printf("Error creating groups view: %v\n", errg)
	}

	if errg != nil {
		return errors.New("Error creating views")
	}

	return nil
}
