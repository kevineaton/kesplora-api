# Kesplora - API

A research and training tool experiment. This is the API for the platform. When completed, the vision is that Kesplora is a self-hosted, open source tool to allow researchers to set up and conduct research on their own platform, including surveys, training, and reports. This is not designed as a marketing or product research tool but is instead specifically geared towards academic research.

This project sprung out of a doctoral research project that had unique needs. Those needs may not be relevant to everyone.

## Status

Very early stages. Structs, funcs, flows, and APIs will all change. Do not use this at this time! If you would like to assist, definitely reach out.

## Major Concepts

Once installed, an administrative `User` can configure the site as the admin. The `Site` is the installation, although multiple instances of the API can be a part of a site's installation (for example, for load balancing).

Once configured, the primary grouping of "stuff" is a `Project`. A `Project` is the primary interaction mechanism for `Participants` on a `Site`.

A `Project` is configured with `Flows` that connect `Modules`. A `Module` consists of `Blocks`. A `Block` is something that a `Participant` will do in a research study. Initially, this will include:

- `Sign Up` for post-consent `Participant` access creation
- `Survey` to gather responses
- `Presentation` for either `Video`, `PDF`, `Download`, or `External` resources the `Participant` will complete.

Since `Consent` is critical, it exists separate from a `Module` and will often be the first thing in a `Flow`.

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

The API supports `access_token` and `refresh_token`. The access_token is short lived and, once expired, a new on can be generated with a refresh_token. The refresh_token is provided in a cookie. However, not all clients can and do support cookies for the calls, so we also support providing the access_token as the `Authorization: Bearer TOKEN` authorization method. In this flow, the `access_token` is provided and a 401 is returned if it is expired. If expired, the call to refresh the token should be made and the call re-tried.

## Contributing

Since this project targets a very specific need, it shouldn't be viewed as a "fit as many features in as possible" project. It's best to raise an issue or send a message prior to taking on any new feature development, unless there is a bug fix in something already developed.

### Code Style

Generally, we use the following casing:

- Filenames are `snake_case` EXCEPT when the tooling requires otherwise (such as Taskfile, etc)
- Variables are `camelCase`
- JSON and ENUMs as `snake_case`
- Structs and functions are either `TitleCase` or `camelCase` depending on visibility

All dates and times will assume UTC. It is up to the client apps to render it in local time if desired.

### Tools

- Task: Used in a similar matter to `make`. See `Taskfile.yml`
- Migrate: Used to handle DB schema migrations
