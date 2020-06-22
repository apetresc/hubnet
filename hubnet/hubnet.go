package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"path"

	"github.com/apetresc/hubnet/backend"
	"github.com/dustin/go-nntp/server"
	_ "github.com/mattn/go-sqlite3"
	gap "github.com/muesli/go-app-paths"
)

func maybefatal(err error, f string, a ...interface{}) {
	if err != nil {
		log.Fatalf(f, a...)
	}
}

func main() {
	log.Printf("Starting up Hubnet...")
	dir, err := gap.NewScope(gap.User, "hubnet").DataDirs()
	maybefatal(err, "Error accessing data dir for hubnet")
	err = os.MkdirAll(dir[0], os.ModeDir)
	maybefatal(err, "Error creating data dir for hubnet")
	db, err := sql.Open("sqlite3", path.Join(dir[0], "hubnet.db"))
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
