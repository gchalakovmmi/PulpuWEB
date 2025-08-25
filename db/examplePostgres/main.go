package main

import (
	"log"
	"net/http"
	"context"
	"fmt"

	"github.com/gchalakovmmi/PulpuWEB/db"
	"github.com/jackc/pgx/v5"
)

func main() {
	dbConnectionDetails, err := db.GetPostgresConfig()
	if err != nil {
			log.Fatalf("Failed to get Postgres config: %v", err)
	}

	http.HandleFunc("/", db.WithDB(dbConnectionDetails, func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		var version string
		err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "<h1>Database Connection Successful</h1><pre>%s</pre>", version)
	}))

	log.Println("Server running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
