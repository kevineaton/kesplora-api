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

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	Yes      = "yes"
	No       = "no"
	SortASC  = "ASC"
	SortDESC = "DESC"
)

type apiConfig struct {
	Environment      string
	APIPort          string
	LogLevelOutput   string
	RootAPIDomain    string
	JWTSigningString string
	SiteCode         string // needed if the site is pending and a new install
	APILevel         string // one of all, admin, participant; used to mount routes

	DBConnection *sqlx.DB
	CacheClient  *redis.Client
	AWSS3Client  *s3.Client
	AWSS3Bucket  string
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
	config.APILevel = envHelper("KESPLORA_API_LEVEL", "all")

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

	// S3
	s3Access := envHelper("KESPLORA_API_S3_ACCESS", "")
	s3Secret := envHelper("KESPLORA_API_S3_SECRET", "")
	s3Bucket := envHelper("KESPLORA_API_S3_BUCKET", "")
	if s3Access != "" && s3Secret != "" && s3Bucket != "" {
		// configure the client
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Access, s3Secret, "")),
		)
		if err != nil {
			fmt.Printf("\n%+v\n", err)
		} else {
			config.AWSS3Client = s3.NewFromConfig(cfg)
			config.AWSS3Bucket = s3Bucket
		}
	}

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

	// set up the routes applicable to everyone
	// We don't mirror these in case we wanted duplicated routes (for example /site vs /participant/site vs /admin/site, which could all return different info)
	r.Get("/", routeApiStatusReady)

	// sites and unauthed admin routes for setup
	r.Get("/site", routeAllGetSite)
	r.Get("/setup", routeAllGetSiteConfiguration)

	// this one is unique as it is un-authed, but mainly for admins
	r.Post("/setup", routeAllConfigureSite)

	// some project and consent routes are available to everyone
	r.Get("/projects", routeAllGetProjects)
	r.Get("/projects/{projectID}", routeAllGetProject)
	r.Get("/projects/{projectID}/consent", routeAllGetConsentForm)
	r.Post("/projects/{projectID}/consent/responses", routeAllCreateConsentResponse)

	// users
	r.Post("/login", routeAllUserLogin)
	r.Post("/logout", routeAllUserLogout)
	r.Get("/me", routeAllGetUserProfile)
	r.Patch("/me", routeAllUpdateUserProfile)
	r.Post("/me/refresh", routeAllUserRefreshAccess)

	//
	// Admin Routes
	//

	// all of these will automatically check if the user is an admin, so that specific check can be ignored
	// in the routes, although there is minimal harm as most require the admin user validity and information
	// anyway
	if config.APILevel == "all" || config.APILevel == "admin" {
		r.Route("/admin", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
						MustBeAdmin:     true,
						ShouldSendError: true,
					})
					if !results.IsAdmin || !results.IsValid {
						return
					}
					ctx := r.Context()
					if results.Site != nil {
						ctx = context.WithValue(r.Context(), appContextSite, results.Site)
					}
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// site
			r.Patch("/site", routeAdminUpdateSite)

			// users
			r.Get("/users", routeAdminGetUsersOnPlatform)
			r.Get("/users/{userID}", routeAdminGetUserOnPlatform)
			r.Get("/users/{userID}/projects", routeAdminGetProjectsForUser)
			r.Get("/users/{userID}/projects/{projectID}", routeAdminGetProjectForUser)
			r.Post("/users/{userID}/projects/{projectID}", routeAdminLinkUserAndProject) // used for overriding, but should be careful due to consent flows
			r.Delete("/users/{userID}/projects/{projectID}", routeAdminUnlinkUserAndProject)

			// projects
			r.Post("/projects", routeAdminCreateProject)
			r.Get("/projects", routeAdminGetProjects)
			r.Get("/projects/{projectID}", routeAdminGetProject)
			r.Patch("/projects/{projectID}", routeAdminUpdateProject)

			// project consent forms
			r.Post("/projects/{projectID}/consent", routeAdminSaveConsentForm)
			r.Delete("/projects/{projectID}/consent", routeAdminDeleteConsentForm)
			r.Get("/projects/{projectID}/consent/responses", routeAdminGetConsentResponses)
			r.Get("/projects/{projectID}/consent/responses/{responseID}", routeAdminGetConsentResponse)
			r.Delete("/projects/{projectID}/consent/responses/{responseID}", routeAdminDeleteConsentResponse)

			// project / users; note that the link and unlink are duplicated routes for convenience
			r.Get("/projects/{projectID}/users", routeAdminGetUsersOnProject)
			r.Post("/projects/{projectID}/users/{userID}", routeAdminLinkUserAndProject) // used for overriding, but should be careful due to consent flows
			r.Delete("/projects/{projectID}/users/{userID}", routeAdminUnlinkUserAndProject)

			// modules, which includes flows
			r.Post("/modules", routeAdminCreateModule)
			r.Get("/modules", routeAdminGetAllSiteModules)
			r.Get("/modules/{moduleID}", routeAdminGetModuleByID)
			r.Patch("/modules/{moduleID}", routeAdminUpdateModule)
			r.Delete("/modules/{moduleID}", routeAdminDeleteModule)

			// project / module links
			r.Get("/projects/{projectID}/flow", routeAdminGetModulesOnProject)
			r.Delete("/projects/{projectID}/flow", routeAdminUnlinkAllModulesFromProject)
			r.Put("/projects/{projectID}/modules/{moduleID}/order/{order}", routeAdminLinkModuleAndProject)
			r.Delete("/projects/{projectID}/modules/{moduleID}", routeAdminUnlinkModuleAndProject)

			// blocks
			r.Get("/blocks", routeAdminGetBlocksOnSite)
			r.Post("/blocks/{blockType}", routeAdminCreateBlock)
			r.Get("/blocks/{blockID}", routeAdminGetBlock)
			r.Patch("/blocks/{blockID}", routeAdminUpdateBlock)
			r.Delete("/blocks/{blockID}", routeAdminDeleteBlock)
			r.Get("/modules/{moduleID}/blocks", routeAdminGetBlocksForModule)
			r.Delete("/modules/{moduleID}/blocks", routeAdminUnlinkAllBlocksFromModule)
			r.Put("/modules/{moduleID}/blocks/{blockID}/order/{order}", routeAdminLinkBlockAndModule)
			r.Delete("/modules/{moduleID}/blocks/{blockID}", routeAdminUnlinkBlockAndModule)

			// user / block progress and info
			r.Put("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/status", notImplementedRoute)
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/status", notImplementedRoute)

			// submissions
			r.Get("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/submissions", routeAdminGetUserSubmissions)
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/submissions", routeAdminDeleteUserSubmissions)
			r.Get("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/submissions/{submissionID}", routeAdminGetUserSubmission)
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/users/{userID}/submissions/{submissionID}", routeAdminDeleteUserSubmission)

			// files
			r.Post("/files", routeAdminUploadFile) // multipart-form
			r.Get("/files", routeAdminGetFiles)
			r.Post("/files/{fileID}", routeAdminReplaceFile)
			r.Delete("/files/{fileID}", routeAdminDeleteFile)
			r.Patch("/files/{fileID}", routeUpdateFileMetadata)
			r.Get("/files/{fileID}", routeAdminGetFileMetaData)
			r.Get("/files/{fileID}/download", routeAdminDownloadFile)

			// notes; NOTE: these are duplicated to allow participants and admins to journal as needed with same routes
			r.Get("/notes", routeAllGetMyNotes)
			r.Post("/notes", routeAllCreateNote)
			r.Get("/notes/{noteID}", routeAllGetMyNoteByID)
			r.Patch("/notes/{noteID}", routeAllUpdateNoteByID)
			r.Delete("/notes/{noteID}", routeAllDeleteMyNoteByID)

		})
	}

	if config.APILevel == "all" || config.APILevel == "participant" {
		r.Route("/participant", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
						ShouldSendError:   true,
						MustBeParticipant: true,
					})
					if !results.IsValid {
						return
					}
					ctx := r.Context()
					if results.Site != nil {
						ctx = context.WithValue(r.Context(), appContextSite, results.Site)
					}
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// projects
			r.Get("/projects", routeParticipantGetProjects)
			r.Delete("/projects/{projectID}", routeParticipantUnlinkUserAndProject)
			r.Get("/projects/{projectID}", routeParticipantGetProject)
			r.Get("/projects/{projectID}/flow", routeParticipantGetProjectFlow)
			r.Get("/projects/{projectID}/consent/responses/{responseID}", routeParticipantGetConsentResponse)
			r.Delete("/projects/{projectID}/consent/responses/{responseID}", routeParticipantDeleteConsentResponse)

			r.Get("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}", routeParticipantGetBlock)
			r.Put("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/status/{status}", routeParticipantSaveBlockStatus)

			// special form routes
			r.Post("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/submissions", routeParticipantSaveFormResponse)
			r.Get("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/submissions", routeParticipantGetFormSubmissions)
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/submissions", routeParticipantDeleteSubmissions)
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/submissions/{submissionID}", routeParticipantDeleteSubmission)

			// participant's can reset their status
			r.Delete("/projects/{projectID}/modules/{moduleID}/blocks/{blockID}/status", routeParticipantRemoveBlockStatus)
			r.Delete("/projects/{projectID}/modules/{moduleID}/status", routeParticipantRemoveBlockStatus)
			r.Delete("/projects/{projectID}/status", routeParticipantRemoveBlockStatus)

			// files
			r.Get("/files/{fileID}", routeParticipantGetFileMetaData)
			r.Get("/files/{fileID}/download", routeParticipantDownloadFile)

			// notes; NOTE: these are duplicated to allow participants and admins to journal as needed with same routes
			r.Get("/notes", routeAllGetMyNotes)
			r.Post("/notes", routeAllCreateNote)
			r.Get("/notes/{noteID}", routeAllGetMyNoteByID)
			r.Patch("/notes/{noteID}", routeAllUpdateNoteByID)
			r.Delete("/notes/{noteID}", routeAllDeleteMyNoteByID)
		})
	}
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
		// to avoid issues with things like terminal prompts, replace !
		config.JWTSigningString = strings.ReplaceAll(config.JWTSigningString, "!", "0")
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
