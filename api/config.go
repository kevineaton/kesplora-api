package api

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
)

var config *apiConfig = nil

const (
	Yes = "yes"
	No  = "no"
)

type apiConfig struct {
	Environment      string
	APIPort          string
	RootAPIDomain    string
	JWTSigningString string
	SiteCode         string // needed if the site is pending and a new install
	DBConnection     *sqlx.DB
}

// SetupConfig is a call to configure the basic required configuration options for the API
func SetupConfig() *apiConfig {
	rand.Seed(time.Now().UnixNano())

	if config != nil {
		return config
	}

	config = &apiConfig{}
	config.Environment = envHelper("KESPLORA_ENVIRONMENT", "test")
	config.APIPort = envHelper("KESPLORA_API_API_PORT", "8080")
	config.RootAPIDomain = envHelper("KESPLORA_DOMAIN", "localhost")
	config.JWTSigningString = envHelper("KESPLORA_JWT_SIGNING", "")

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
	r.Use(middleware.RealIP) // TODO: make this optional
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// access token middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			found := false
			user := jwtUser{}
			expired := false

			// first cookie, then authorization
			accessCookie, err := r.Cookie(tokenTypeAccess)
			if err == nil && accessCookie != nil {
				// try parsing
				user, err = parseJWT(accessCookie.Value)
				if err == nil && user.ID != 0 {
					found = true
				}

			}

			if !found {
				// check the header
				access := r.Header.Get("Authorization")
				if strings.HasPrefix(access, "Bearer") {
					parts := strings.Split(access, " ")
					if len(parts) > 0 {
						access = parts[1]
					}
				}
				user, err = parseJWT(access)
				if err == nil && user.ID != 0 {
					found = true
				}
			}

			if found {
				// check if expired
				expiresAt, _ := time.Parse("2006-01-02T15:04:05Z", user.Expires)
				if expiresAt.Before(time.Now()) {
					expired = true
				}
			}

			ctx := context.WithValue(r.Context(), appContextKeyFound, found)
			ctx = context.WithValue(ctx, appContextKeyUser, user)
			ctx = context.WithValue(ctx, appContextKeyExpired, expired)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	cors := cors.New(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true // TODO: we probably want to let the setup set this in the config
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-TOKEN", "RANGE", "ACCEPT-RANGE", "Access-Conrol-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	// set up a Not Implemented handler just as a placeholder
	notImplementedRoute := func(w http.ResponseWriter, r *http.Request) {
		sendAPIError(w, api_error_not_implemented, errors.New("not implemented"), map[string]string{})
	}

	// set up the routes
	r.Get("/", routeApiStatusReady)

	r.Get("/setup", routeGetSiteConfiguration)
	r.Post("/setup", routeConfigureSite)

	// sites
	r.Get("/site", routeGetSite)
	r.Patch("/site", routeUpdateSite)

	// users
	r.Post("/login", routeUserLogin)
	r.Get("/me", routeGetUserProfile)
	r.Patch("/me", routeUpdateUserProfile)
	r.Post("/me/refresh", notImplementedRoute)

	// projects
	r.Post("/projects", notImplementedRoute)
	r.Get("/projects", notImplementedRoute)
	r.Get("/projects/{projectID}", notImplementedRoute)
	r.Patch("/projects/{projectID}", notImplementedRoute)

	// project / users
	r.Get("/projects/{projectID}/users", notImplementedRoute)
	r.Post("/projects/{projectID}/users/{userID}", notImplementedRoute)

	// modules, which includes flows
	r.Post("/modules", notImplementedRoute)
	r.Get("/modules", notImplementedRoute)
	r.Get("/modules/{moduleID}", notImplementedRoute)
	r.Patch("/modules/{moduleID}", notImplementedRoute)
	r.Delete("/modules/{moduleID}", notImplementedRoute)
	r.Put("/projects/{projectID}/modules/{moduleID}/{order}", notImplementedRoute)
	r.Delete("/projects/{projectID}/modules/{moduleID}/{order}", notImplementedRoute)

	// blocks
	r.Post("/blocks", notImplementedRoute)
	r.Get("/blocks", notImplementedRoute)
	r.Get("/blocks/{blockID}", notImplementedRoute)
	r.Patch("/blocks/{blockID}", notImplementedRoute)
	r.Delete("/blocks/{blockID}", notImplementedRoute)
	r.Put("/modules/{moduleID}/blocks/{blockID}/{order}", notImplementedRoute)
	r.Delete("/modules/{moduleID}/{blockID}/{order}", notImplementedRoute)

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

func CheckConfiguration() {
	// this should check the db, make sure things are good to go
	// since the DB would have nuked before here, check if there's any users or site info
	site, err := GetSite()
	if err != nil || site.Status == "pending" {
		// if not, show a code that allows a user to initiate the site
		code := randomString(32)
		config.SiteCode = code
		fmt.Println("")
		fmt.Printf("-------------------------------------------------------------------\n")
		fmt.Printf("-- Your site is not configured, see the output below             --\n")
		fmt.Printf("--  Site Code:     %s              --\n", code)
		fmt.Printf("--                                                               --\n")
		fmt.Printf("-- Why am I seeing this?                                         --\n")
		fmt.Printf("-- The DB you supplied does not have an active site              --\n")
		fmt.Printf("-- so you must configure it with the above code and              --\n")
		fmt.Printf("-- the chosen client pointed at the API. See the docs.           --\n")
		fmt.Printf("-------------------------------------------------------------------\n")
	}

	if config.JWTSigningString == "" {
		// probably a bad day, but we won't block it; we will want to output it, especially in multi-host installs
		config.JWTSigningString = randomString(32)
		fmt.Println("")
		fmt.Printf("-------------------------------------------------------------------\n")
		fmt.Printf("-- JWT Signing Key Generated: %s   --\n", config.JWTSigningString)
		fmt.Printf("--                                                               --\n")
		fmt.Printf("-- Why am I seeing this: No KESPLORA_JWT_SIGNING environment     --\n")
		fmt.Printf("-- variable was provided so we generated a new one for you.      --\n")
		fmt.Printf("-- You will need to capture this for future server installations.--\n")
		fmt.Printf("-------------------------------------------------------------------\n")
	}

}
