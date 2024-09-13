# Mattermost Bot Reminder

## Prerequisites

- go
- golang-migrate (See [Migrations](#migrations))
- docker
- docker-compose

## Installation

1. Copy `.env.example` to `.env` and fill in all the required information
2. You have to decide to use program either for test or not (see differences in [Container description](#container-description) section)
   1. Test: run `docker-compose --profile test up -d`
   2. Prod: run `docker-compose up -d`
3. Run `migrate -database DB_URL -source file://migrations up` (see [Migrations](#migrations) for more information)
4. Follow [this](https://developers.mattermost.com/integrate/slash-commands/custom/) tutorial to add slash commands. As a url use POST `http://<host>:8080/mattermost/reminders` (for test `<host>` is `test_mm`)
5. Follow [this](https://developers.mattermost.com/integrate/webhooks/incoming/) tutorial to add an incoming webhook
6. Add the webhook URL to the `.env` file as a `MM_IN_HOOK` (more likely you'll have to restart the `poller` service to apply changes)

## Usage

Bot has several commands:

- `add NAME CRON-RULE` - creates new reminder with the specified rule
- `list` - shows the reminders table with their names, rules and channels

`CRON-RULE` could be set with variable amount of values, but recommended and tested is 7.

CRON-RULE: "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"

### Examples

Rule will create weekly reminder that will be triggered at 12:00 on fridays repeatedly

```text
/reminder create weekly_reminder "0 0 12 * * FRI/1 *"
```

## Configuration

Database: MySQL

### Container description

1. `db` - mysql database
2. `reminder` - server with the main logic which handles a mattermost slash commands and provides useful API endpoints
3. `poller` simple service that periodically polls the `reminder` container for reminds and sends them to a corresponding mattermost channel using webhook
4. `test_mm` test profile - container that holds a test local mattermost server

## Migrations

Migrations could be done using [this](https://github.com/golang-migrate/migrate) tool.

To install it, run:

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

(This command will install `migrate` cli program into `$HOME/go/bin` directory, maybe you will have to add this directory to the $PATH)

To create migration:

```bash
migrate create -ext sql -dir migrations -seq MIGRATION_NAME
```

To run migrations:

```bash
migrate -database 'mysql://MYSQL_USER:MYSQL_PASSWORD@tcp(HOST:PORT)/NAME' -source file://migrations up
```

All the uppercased from the command above you specify in your `.env` file. Using `.env.example` you get:

```bash
migrate -database 'mysql://reminders:XXXXXX@tcp(db:3306)/reminders' -source file://migrations up
```
