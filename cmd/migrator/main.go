package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		cmd   = flag.String("cmd", "up", "migrations command: up, down")
		dir   = flag.String("dir", "migrations", "path to migrations directory")
		steps = flag.Int("steps", 1, "number of migration steps")
	)
	flag.Parse()

	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		log.Fatal("POSTGRES_URL variable is required")
	}

	m, err := migrate.New("file://"+*dir, url)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	defer m.Close()

	switch *cmd {
	case "up":
		log.Println("running migrations up...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to run migrations up: %v", err)
		}
		log.Println("migrations completed successfully")
	case "down":
		log.Printf("rolling back %d migration(s)...", *steps)
		if err := m.Steps(*steps); err != nil {
			log.Fatalf("failed to rollback migrations: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s, use: up, down", *cmd)
	}

}
