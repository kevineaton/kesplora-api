package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
)

var config *apiConfig = nil

type apiConfig struct {
	APIPort      string
	DBConnection *sqlx.DB
}

// SetupConfig is a call to configure the basic required configuration options for the API
func SetupConfig() *apiConfig {
	rand.Seed(time.Now().UnixNano())

	if config != nil {
		return config
	}

	config = &apiConfig{}
	config.APIPort = envHelper("KESPLORA_API_API_PORT", "8080")

	// now we ensure we can connect to the DB
	dbConnectionString := envHelper("KESPLORA_API_DB_CONNECTION", "root:password@tcp(localhost:3306)/Kesplora")
	conn, err := sqlx.Open("mysql", dbConnectionString)
	if err != nil {
		panic(err)
	}
	conn.SetMaxIdleConns(100)

	// make sure queries hit
	_, err = conn.Exec("set session time_zone='-0:00'")
	maxTies := 10
	secondsBetweenTries := 5
	if err != nil {
		// try again until we decide we can try no longer
		fmt.Printf("could not connect to db: %+v\ntrying again in %d seconds\n", err, secondsBetweenTries)
		for i := 1; i <= maxTies; i++ {
			time.Sleep(time.Duration(secondsBetweenTries) * time.Second)
			_, err = conn.Exec("set session time_zone='-0:00'")
			if err == nil {
				break
			}
			if i == maxTies {
				panic("could not connect to the MySQL database, shutting down")
			}
		}
	}
	config.DBConnection = conn

	return config
}

var r *chi.Mux

// SetupAPI sets up an API Mux for handling the calls
func SetupAPI() *chi.Mux {
	if r != nil {
		return r
	}
	r = chi.NewRouter()

	// configure our middlewares here
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Types", "X-CSRF-TOKEN", "RANGE", "ACCEPT-RANGE"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	// set up a Not Implemented handler just as a placeholder
	notImplementedRoute := func(w http.ResponseWriter, r *http.Request) {
		sendAPIError(w, api_error_not_implemented, nil)
	}

	// set up the routes
	r.Get("/", routeApiStatusReady)

	r.Get("/setup", notImplementedRoute)
	r.Post("/setup", notImplementedRoute)

	return r
}

// envHelper is a simple helper to check the env for a value or return a default
func envHelper(key, defaultValue string) string {
	found := os.Getenv(key)
	if found != "" {
		return found
	}
	return defaultValue
}
