package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/apetresc/hubnet/backend"
	"github.com/dustin/go-nntp/server"
	_ "github.com/mattn/go-sqlite3"
)

func maybefatal(err error, f string, a ...interface{}) {
	if err != nil {
		log.Fatalf(f, a...)
	}
}

func main() {
	log.Printf("Starting up Hubnet...")
	db, err := sql.Open("sqlite3", "./hubnet.db")
	maybefatal(err, "Error connecting to database: %s", err)
	defer db.Close()

	backend.EnsureViews(db)
	backend := backend.SQLBackend{
		DB: db,
	}

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
