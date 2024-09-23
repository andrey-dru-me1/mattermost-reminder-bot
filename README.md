# Mattermost Bot Reminder

## Prerequisites

- go
- golang-migrate (See [Migrations](#migrations))
- docker
- docker-compose

## Installation

### Test

1. (Optional) If you want to test reminder bot on a local mattermost server, you should run `docker-compose up test_mm -d`
   1. If you decide to test the bot like this, you could access mattermost either from browser on <http://localhost:8065> or from desktop client adding a server with the same url
2. Follow [this](https://developers.mattermost.com/integrate/webhooks/incoming/) tutorial to add an incoming webhook to a mattermost server
3. Follow [this](https://developers.mattermost.com/integrate/slash-commands/custom/) tutorial to add slash commands to a mattermost server. As a url use POST `http://<host>:8080/mattermost/reminders` (for local server `<host>` is `reminder`)
   1. Probably you will have to add `<host>` to an 'Untrusted internal connections' in mattermost server. To do so, go to `System Console/Environment/Developer` and add your `<host>` to the `Allow untrusted internal connections to:` field.
4. Copy `.env.example` to `.env` and fill in all the required information (see [this section](#env-variables) if you got confused about some variables)
   1. Note, that you most likely want to replace a `localhost` in `MM_IN_HOOK` variable with the host your mattermost server runs at (`test_mm` for local mattermost server)
5. Run `docker-compose up db -d`
6. Run `migrate -database DB_URL -source file://migrations up` (see [Migrations](#migrations) for more information)
7. Run `docker-compose up -d`

### Production

#### Dockerhub

You could use pre-built reminder image from [dockerhub](https://hub.docker.com/repository/docker/andreydrumel/mm-remind-bot/general) as such:

```yaml
# docker-compose.yaml
services:
   reminder:
      image: andreydrumel/mm-remind-bot:1.0.0
      ...
```

```Dockerfile
# Dockerfile
FROM andreydrumel/mm-remind-bot:1.0.0
...
```

Do not forget to provide all the necessary ENV's! They are listed in [Container desription](#container-description) section.

Poller service is not provided in dockerhub. You can either write this service on your own (it's not so hard), use low-code solution (such as n8n) or [build your own image from source](#build-from-source).

#### Build from source

There are Dockerfiles in both `poller` and `reminder` root subdirectories, so you could go in there and run `docker build .`.

Otherwise you could use `docker-compose.prod.yaml` from root directory either as a template or the actual compose-file.

## Usage

Bot has several commands:

- `add, create NAME CRON_RULE MESSAGE` - creates new reminder
- `list, ls` - lists all reminders
- `delete, del, remove, rm ID...` - deletes a reminders with ID... identifiers
- `timezone, tz LOCATION` - updates channel timezone
- `timezone, tz` - shows current location

`CRON-RULE`:

- "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"
- "Minutes Hours DayOfMonth Month DayOfWeek Year" (Seconds default to 0)
- "Minutes Hours DayOfMonth Month DayOfWeek" (Year defaults to *)
- Month: 1-12 or JAN-DEC
- DayOfWeek 0-7 or SUN-SAT (both 0 and 7 stand for SUN)
- `*` - any value (`0 12 * * *` - 12:00 every day every month every year)
- `/` - time period (`*/5 * * * *` - every 5 minute of every day every month every year)
- `,` - list separator (`0 12 10,25 * *` - 12:00 every 10th and 25th day of every month)
- `-` - range (`0 12 * MON-FRI *` - 12:00 every workday)
- `L` - last (`0 12 * 5L *` - 12:00 last friday every month)
- `#` - numbered (`0 12 * TUE#2 *` - 12:00 second tuesday of every month)
- More information: <https://github.com/gorhill/cronexpr?tab=readme-ov-file>

`LOCATION`: TZ identifier (for example `Asia/Novosibirsk`)

### Examples

Rule will create weekly reminder that will be triggered at 12:00 on fridays repeatedly

```text
/reminder add "Weekly Reminder" "0 0 12 * * FRI *" "Another week is coming to an end!"
```

---

Rule will create a reminder that triggers at 9:00, 12:00, 15:00 and 18:00 every monday and thursday

```text
/reminder add "Repeatedly reminder" "0 9-18/3 * * MON,THU"
```

## Configuration

Database: MySQL

### Env variables

- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `DB_HOST` - DataBase Host - host to access database, defaults to `db` service
- `DB_PORT` - DataBase Port - mysql default is `3306`, but you can change it here
- `DB_NAME` - DataBase Name - default is `reminders`, but if you want to use another name, you should rename it here
- `MM_IN_HOOK` - MatterMost Incoming Webhook - a URL that you receive after [creating incoming web hook](https://developers.mattermost.com/integrate/webhooks/incoming/)
- `MM_SC_TOKEN` - MatterMost Slash Command Token - token that you receive after [creating slash command](https://developers.mattermost.com/integrate/slash-commands/custom/)
- `DEFAULT_TZ` - Default Time Zone - for the whole mattermost server

### Container description

1. `db` - mysql database
   1. `MYSQL_USER`
   2. `MYSQL_PASSWORD`
   3. `DB_NAME`
2. `reminder` - server with the main logic which handles a mattermost slash commands and provides useful API endpoints
   1. `MYSQL_USER`
   2. `MYSQL_PASSWORD`
   3. `DB_HOST`
   4. `DB_PORT`
   5. `DB_NAME`
   6. `MM_SC_TOKEN` - to verify the server attempting to use this command
   7. `DEFAULT_TZ`
3. `poller` simple service that periodically polls the `reminder` container for reminds and sends them to a corresponding mattermost channel using webhook
   1. `MM_IN_HOOK` - to send messages once reminders trigger
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
