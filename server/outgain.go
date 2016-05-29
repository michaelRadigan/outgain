package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/egnwd/outgain/server/engine"
	"github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/egnwd/outgain/server/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := openDb()

	fmt.Println("%s, %s", db, err)
	

	staticDir := flag.String("static-dir", "client/dist", "")
	redirectPlainHTTP := flag.Bool("redirect-plain-http", false, "")
	flag.Parse()

	engine := engine.NewEngine()

	handler := routes.GetHandler(*staticDir, engine)
	if *redirectPlainHTTP {
		handler = redirectPlainHTTPMiddleware(handler)
	}

	go engine.Run()

	log.Printf("Listening on port %s", port)
	http.ListenAndServe(":"+port, handler)
}

func redirectPlainHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-forwarded-proto") != "https" {
			hostname := strings.Split(r.Host, ":")[0]
			redirectTo := fmt.Sprintf("https://%s%s", hostname, r.URL.String())
			http.Redirect(w, r, redirectTo, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func openDb() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	connection, _ := pq.ParseURL(url)
	connection += " sslmode=require"

	db, err := sql.Open("postgres", connection)
	if err != nil {
		log.Println(err)
	}

	return db, err
}
