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
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var config *apiConfig = nil

const (
	Yes = "yes"
	No  = "no"
)

type apiConfig struct {
	Environment      string
	APIPort          string
	LogLevelOutput   string
	RootAPIDomain    string
	JWTSigningString string
	SiteCode         string // needed if the site is pending and a new install
	DBConnection     *sqlx.DB
	CacheClient      *redis.Client
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

	config.LogLevelOutput = strings.ToUpper(envHelper("KESPLORA_LOG_LEVEL", "WARN"))
	log.SetFormatter((&log.JSONFormatter{}))
	switch config.LogLevelOutput {
	case LogLevelTrace:
		log.SetLevel(log.TraceLevel)
	case LogLevelDebug:
		log.SetLevel(log.DebugLevel)
	case LogLevelInfo:
		log.SetLevel(log.InfoLevel)
	case LogLevelWarn:
		log.SetLevel(log.WarnLevel)
	case LogLevelError:
		log.SetLevel(log.ErrorLevel)
	case LogLevelFatal:
		log.SetLevel(log.FatalLevel)
	case LogLevelPanic:
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	// now we ensure we can connect to the DB
	dbConnectionString := envHelper("KESPLORA_API_DB_CONNECTION", "root:password@tcp(localhost:3306)/Kesplora")
	conn, err := sqlx.Open("mysql", dbConnectionString)
	if err != nil {
		panic(err)
	}
	conn.SetMaxIdleConns(100)

	// make sure queries hit
	_, err = conn.Exec("set session time_zone='-0:00'")
	maxTries := 10
	secondsBetweenTries := 5
	if err != nil {
		// try again until we decide we can try no longer
		fmt.Printf("could not connect to db: %+v\ntrying again in %d seconds\n", err, secondsBetweenTries)
		for i := 1; i <= maxTries; i++ {
			time.Sleep(time.Duration(secondsBetweenTries) * time.Second)
			_, err = conn.Exec("set session time_zone='-0:00'")
			if err == nil {
				break
			}
			if i == maxTries {
				panic("could not connect to the MySQL database, shutting down")
			}
		}
	}
	config.DBConnection = conn

	// now the cache

	cacheAddress := envHelper("KESPLORA_API_CACHE_ADDRESS", "localhost:6379")
	cachePassword := envHelper("KESPLORA_API_CACHE_PASSWORD", "")
	config.CacheClient = redis.NewClient(&redis.Options{
		Addr:     cacheAddress,
		Password: cachePassword,
		DB:       0,
	})
	_, err = config.CacheClient.Ping().Result()
	if err != nil {
		maxTries = 10
		for i := 1; i <= maxTries; i++ {
			fmt.Printf("\n Cache Error, this is attempt %d of %d. Waiting %d seconds...\n", i, maxTries, secondsBetweenTries)
			fmt.Printf("\n\t %s\n", cacheAddress)
			time.Sleep(time.Duration(secondsBetweenTries) * time.Second)
			_, err = config.CacheClient.Ping().Result()
			if err == nil {
				break
			}
			if i == maxTries {
				panic("Could not connect to the cache server, shutting down")
			}
		}
	}
	config.CacheClient.FlushAll().Result()

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
				if access != "" {
					user, err = parseJWT(access)
					if err == nil && user.ID != 0 {
						found = true
					}
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
	r.Post("/logout", routeUserLogout)
	r.Get("/me", routeGetUserProfile)
	r.Patch("/me", routeUpdateUserProfile)
	r.Post("/me/refresh", routeUserRefreshAccess)

	// projects
	r.Post("/projects", routeCreateProject)
	r.Get("/projects", routeGetProjects)
	r.Get("/projects/{projectID}", routeGetProject)
	r.Patch("/projects/{projectID}", routeUpdateProject)

	// project consent forms
	r.Post("/projects/{projectID}/consent", routeSaveConsentForm)
	r.Get("/projects/{projectID}/consent", routeGetConsentForm)
	r.Delete("/projects/{projectID}/consent", routeDeleteConsentForm)
	r.Post("/projects/{projectID}/consent/responses", routeCreateConsentResponse)
	r.Get("/projects/{projectID}/consent/responses", routeGetConsentResponses)
	r.Get("/projects/{projectID}/consent/responses/{responseID}", routeGetConsentResponse)
	r.Delete("/projects/{projectID}/consent/responses/{responseID}", routeDeleteConsentResponse)

	// project / users
	r.Get("/projects/{projectID}/users", notImplementedRoute)
	r.Post("/projects/{projectID}/users/{userID}", routeLinkUserAndProject) // used for overriding, but should be careful due to consent flows
	r.Delete("/projects/{projectID}/users/{userID}", routeUnlinkUserAndProject)

	// modules, which includes flows
	r.Post("/modules", routeCreateModule)
	r.Get("/modules", routeGetAllSiteModules)
	r.Get("/modules/{moduleID}", routeGetModuleByID)
	r.Patch("/modules/{moduleID}", routeUpdateModule)
	r.Delete("/modules/{moduleID}", routeDeleteModule)

	// project / module links
	r.Get("/projects/{projectID}/flow", routeGetModulesOnProject)
	r.Delete("/projects/{projectID}/flow", routeUnlinkAllModulesFromProject)
	r.Put("/projects/{projectID}/modules/{moduleID}/order/{order}", routeLinkModuleAndProject)
	r.Delete("/projects/{projectID}/modules/{moduleID}", routeUnlinkModuleAndProject)

	// blocks
	r.Get("/blocks", routeGetBlocksOnSite)
	r.Post("/blocks/{blockType}", routeCreateBlock)
	r.Get("/blocks/{blockID}", routeGetBlock)
	r.Patch("/blocks/{blockID}", routeUpdateBlock)
	r.Delete("/blocks/{blockID}", routeDeleteBlock)
	r.Get("/modules/{moduleID}/blocks", routeGetBlocksForModule)
	r.Delete("/modules/{moduleID}/blocks", routeUnlinkAllBlocksFromModule)
	r.Put("/modules/{moduleID}/blocks/{blockID}/order/{order}", routeLinkBlockAndModule)
	r.Delete("/modules/{moduleID}/blocks/{blockID}", routeUnlinkBlockAndModule)

	// user / block progress
	r.Put("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/status", notImplementedRoute)
	r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/status", notImplementedRoute)

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

func setupTesting() {
	SetupConfig()
	SetupAPI()
	_, err := GetSite()
	if err != nil {
		err = createTestSite(&Site{
			Status: SiteStatusActive,
		})
		if err != nil {
			panic(err)
		}
	}
}
