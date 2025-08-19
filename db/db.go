package db

import (
	"log"
	"fmt"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"

	"errors"
	"os"
	"strconv"
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

func GetPostgresConfig() (ConnectionDetails, error) {
    // Required environment variables
    user := os.Getenv("POSTGRES_USER")
    if user == "" {
        return ConnectionDetails{}, errors.New("POSTGRES_USER environment variable not set")
    }

    password := os.Getenv("POSTGRES_PASSWORD")
    if password == "" {
        return ConnectionDetails{}, errors.New("POSTGRES_PASSWORD environment variable not set")
    }

    serverIP := os.Getenv("POSTGRES_IP")
    if serverIP == "" {
        return ConnectionDetails{}, errors.New("POSTGRES_IP environment variable not set")
    }

    schema := os.Getenv("POSTGRES_DB")
    if schema == "" {
        return ConnectionDetails{}, errors.New("POSTGRES_DB environment variable not set")
    }

    portStr := os.Getenv("POSTGRES_PORT")
    if portStr == "" {
        return ConnectionDetails{}, errors.New("POSTGRES_PORT environment variable not set")
    }
                                                                                                
    port, err := strconv.Atoi(portStr)                                                                                            
    if err != nil {                                                                                                               
        return ConnectionDetails{}, errors.New("invalid POSTGRES_PORT format")                                                                      
    }                                                                                                                             
                                                                                                
    return ConnectionDetails{                                                                                                    
        User:     user,                                                                                                             
        Password: password,                                                                                                         
        ServerIP: serverIP,                                                                                                         
        Schema:   schema,                                                                                                           
        Port:     port,                                                                                                             
    }, nil                                                                                                                        
}
