# Kesplora - API

A research and training tool experiment. This is the API for the platform. When completed, the vision is that Kesplora is a self-hosted, open source tool to allow researchers to set up and conduct research on their own platform, including surveys, training, and reports. This is not designed as a marketing or product research tool but is instead specifically geared towards academic research.

This project sprung out of a doctoral research project that had unique needs. Those needs may not be relevant to everyone.

## Status

Very early stages. Structs, funcs, flows, and APIs will all change. Do not use this at this time! If you would like to assist, definitely reach out.

**NOTE**: Since the APIs are fluid and quickly changing, test coverage is below acceptable standards for production use. Use at your own risk, especially now!

## Configuration

Most configuration is handled through the `Config.go` file reading from the environment. Of note, the following values deserve explanation:

- `KESPLORA_ENVIRONMENT` (`test`): One of `test`, `dev`, or `production`. Currently has no impact on the business logic.
- `KESPLORA_API_PORT` (`8080`): The port for the HTTP server to listen on. This is usually served behind proxy, such as nginx, that handles SSL termination
- `KESPLORA_DOMAIN` (`localhost`): The domain the HTTP server listens on. Used for things like HTTP Cookie scoping
- `KESPLORA_JWT_SIGNING` (`will be randomly generated`): The signing key to use for JWT encryption. Will be randomly generated if not provided. This should be consistent across similar deployments.
- `KESPLORA_API_LEVEL` (`all`): One of `all`, `admin`, or `participant`. Which API routes to serve. Useful if you want to restrict the admin routes behind different VPC or firewalls.
- `KESPLORA_API_DB_CONNECTION` (`root:password@tcp(localhost:3306)/Kesplora`): The DB connection string. Currently only MySQL is supported.
- `KESPLORA_API_CACHE_ADDRESS` (`localhost:6379`): The connection string for the Redis server.
- `KESPLORA_API_CACHE_PASSWORD` (``): The password for the Redis connection.
- `KESPLORA_API_S3_ACCESS` (``): The S3 access token
- `KESPLORA_API_S3_SECRET` (``): The S3 secret token
- `KESPLORA_API_S3_BUCKET` (``): The S3 bucket
- `KESPLORA_API_S3_REGION` (`us-east-1`): The S3 region

## Set Up

The current architecture is "one DB, one site". Although the `Sites` table has an ID, the current expectation is a single ID for a single site. This site is checked on startup to determine if the site should be set up or not. If the site's `status` field is `pending`, the configuration will output a code that needs to be sent up when configuring the site in order to make it `active`.

To run an instance, you will need to have the following:

- The Docker images or binaries you want to run, configured to speak with each other
- An AWS IAM with read and write access to an S3 bucket (additional providers coming)
- A Mailgun account for sending emails (additional providers coming)

For clients, when the app starts up, a configuration code will be output to the terminal. This is not stored anywhere. A call to GET `/site` will fail, but a call to GET `/setup` will state whether the site is already configured. Nothing will be saved in the DB at this point. The client should make a `POST` to `/setup` to configure the site. This is almost exactly like the call to PATCH `/site` but will also configure the user account and additional data as needed.

So, the TLDR is on first set up, the client will need the configuration code, should GET to `/setup` and then POST to `/setup`.

## Major Concepts

Once installed, an administrative `User` can configure the site as the admin. The `Site` is the installation, although multiple instances of the API can be a part of a site's installation (for example, for load balancing).

Once configured, the primary grouping of "stuff" is a `Project`. A `Project` is the primary interaction mechanism for `Participants` on a `Site`.

A `Project` is configured with `Flows` that connect `Modules`. A `Module` consists of `Blocks`. A `Block` is something that a `Participant` will do in a research study. Initially, this will include:

- `Consent` forms for sign up and management
- `Modules`
- `Blocks`, which currently include:
  - `Text` - Just textual information
  - `Embed` - Embedded files, such as `PDF`s or `Youtube` videos
  - `External` - Links to external sources
  - `File` - A downloadable file
  - `Form` - A form for collectable information, such as surveys and tracking
- `Notes` such as Journal entries or block notes. Participants and Admins can use this to record thoughts or really anything.
- `Reports` for administrators to be able to analyze data, such as participant progress and `Form` responses

Since `Consent` is critical, it exists separate from a `Block`. The consent will be linked directly to a `Project` and must be provided prior to access to the `Project`.

So, for example, you could have a theoretical `Project` that wants to measure data retention on a topic. Your `Project` could have the following `Flow`:

- `Consent`
- `Sign Up` with a participant code and password
- `Survey` for pre-exposure
- `Module` 1 with an introduction
- `Module` 2 with a video presentation and download PDF
- `Module` 3 with a summary PDF
- `Survey` to complete

In the above, the researcher would be able to generate a `Report` of the `Participant` (depending on configuration) activities, results of the surveys, and more.

### Authentication

The API supports `access` and `refresh`. The `access` is short lived and, once expired, a new on can be generated with a `refresh`. The `refresh` is provided in a cookie. However, not all clients can and do support cookies for the calls, so we also support providing the access as the `Authorization: Bearer TOKEN` authorization method. In this flow, the `access` is provided and a 401 is returned if it is expired. If expired, the call to `refresh` the token should be made and the call re-tried.

## Contributing

Since this project targets a very specific need, it shouldn't be viewed as a "fit as many features in as possible" project. It's best to raise an issue or send a message prior to taking on any new feature development, unless there is a bug fix in something already developed.

### Code Style

Generally, we use the following casing:

- Filenames are `snake_case` EXCEPT when the tooling requires otherwise (such as Taskfile, etc)
- Variables are `camelCase`
- JSON properties are `camelCase` to match variables
- ENUMs as `snake_case`
- Structs and functions are either `TitleCase` or `camelCase` depending on visibility

All dates and times will assume UTC. It is up to the client apps to render it in local time if desired.

#### Structs

Structs should have both `json` and `db` tags. At the bottom of every struct, there should be a `processForDB` and `processForAPI` method for the struct that handles things like datetime conversions. See, for example, `api/tokens.go`.

### Tools

- Task: Used in a similar matter to `make`. See `Taskfile.yml`
- Migrate: Used to handle DB schema migrations

## Roadmap

We are still very early in development. If a struct or file has minimal code, it has not been fully thought out and is a place holder. Aside from functionality, there's a few non-functionality improvements:

[ ] Integrate an email system that can use Mailgun at the least (but in a way that supports others?)

[ ] Add Open API Specification 3 files to document the API

[ ] Add CI

[ ] Add hcaptcha for signups and setups
