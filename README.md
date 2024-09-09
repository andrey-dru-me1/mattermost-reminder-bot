# Mattermost Bot Reminder

## Prerequisites

- go
- golang-migrate (See [Migrations](#migrations))
- docker
- docker-compose

## Installation

1. Copy `.env.example` to `.env` and fill in all the database information
2. (Optional) run `docker-compose up -d` to run a containerized mysql server
   1. Otherwise you shoud have another database instance
3. Run `migrate -database DB_URL -source file://migrations up` (See [Migrations](#migrations) for more information)
4. Type `go run .` to run a server
5. Follow [this](https://developers.mattermost.com/integrate/slash-commands/custom/) tutorial to add slash commands. As a url use POST `http://<host>:8080/mattermost/reminders`

## Usage

Bot has several commands:

- `create NAME CRON-RULE` - creates new reminder with the specified rule
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

## Migrations

Migrations could be done using [this](https://github.com/golang-migrate/migrate) tool.

To install it, run:

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

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
migrate -database 'mysql://reminders:XXXXXX@tcp(localhost:3306)/reminders' -source file://migrations up
```
