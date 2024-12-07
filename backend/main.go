package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func handler(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hello %s", r.URL.Path[1:])
}

func main() {
    dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
    fmt.Println("Starting")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))

}
