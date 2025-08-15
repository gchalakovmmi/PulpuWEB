package db

import (
	"log"
	"fmt"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
)

type ConnectionDetails struct {
	User string
	Password string
	ServerIP string
	Port int
	Schema string
}

func WithDB(cd ConnectionDetails, handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cd.User, cd.Password, cd.ServerIP, cd.Port, cd.Schema)
		conn, err := pgx.Connect(context.Background(), url)
		if err != nil {
			log.Println(err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer conn.Close(context.Background())

		handler(w, r, conn)
	}
}
