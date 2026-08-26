// Command biosonar is the entry point for the biosonar seafloor-classification
// service. In normal mode it serves the /api HTTP API; with --smoke-test it
// runs a self-contained persistence/restart check and exits.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"task250-biosonar/internal/httpapi"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func main() {
	var addr, dbPath string
	var smoke bool
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&dbPath, "db", "biosonar.db", "SQLite database path")
	flag.BoolVar(&smoke, "smoke-test", false, "run self-contained smoke test and exit")
	flag.Parse()

	if smoke {
		if err := runSmokeTest(); err != nil {
			fmt.Fprintln(os.Stderr, "smoke-test FAILED:", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test OK")
		os.Exit(0)
	}

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := httpapi.NewServer(svc)
	log.Printf("biosonar listening on %s (db=%s)", addr, dbPath)
	if err := srv.Listen(addr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
