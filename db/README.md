# WithDB
Allows you to connect to a Postgres database in an easy and clean way.
Here is an example:
```go
package main

import(
    "os"
    "log"
    "fmt"
    "error"
    "net/http"
    "github.com/jackc/pgx/v5"
)

// Get connection details from environment variables
func getDBConnectionDetails(){
	dbConnectionDetails := db.ConnectionDetails{
		User: 		os.Getenv("POSTGRES_USER"),
		Password:	os.Getenv("POSTGRES_PASSWORD"),
		ServerIP:	os.Getenv("POSTGRES_CONTAINER_NAME"),
		Schema:		os.Getenv("POSTGRES_DB"),
	}
	var err error
	dbConnectionDetails.Port, err = strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		panic(fmt.Sprintf("Invalid POSTGRES_PORT: %v", err))
	}
    return dbConnectionDetails
}

func main() {
	dbConnectionDetails := getDBConnectionDetails()
    // Define endpoint
	http.HandleFunc("/dbtest", db.WithDB(dbConnectionDetails, func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
		var fld string
		err := conn.QueryRow(ctx, "select 'Hello World!' as fld").Scan(&fld)
		if err != nil {
            log.Println("Example query failed. Error:\n%v")
            http.Error(w, "Database error", http.StatusInternalServerError)
		}
		fmt.Println(fld)
		// templ.Handler(home.Home()).ServeHTTP(w, r) // Render a web page.
	}))
}
```
