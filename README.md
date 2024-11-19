# Mattermost Bot Reminder

This `README` is also available on other languages: [Русский](./README.ru.md)

## Table of contents

- [Mattermost Bot Reminder](#mattermost-bot-reminder)
  - [Table of contents](#table-of-contents)
  - [Description](#description)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
    - [Test](#test)
    - [Production](#production)
      - [Dockerhub](#dockerhub)
      - [Build from source](#build-from-source)
  - [Usage](#usage)
    - [Cron rule](#cron-rule)
    - [Location](#location)
    - [Webhook](#webhook)
    - [Examples](#examples)
  - [Configuration](#configuration)
    - [.env file](#env-file)
    - [Container description](#container-description)
  - [Migrations](#migrations)

## Description

The Reminder Bot is a Mattermost bot that allows you to create and manage both periodic and one-time reminders sent to Mattermost channels. It offers flexible reminder scheduling, the ability to send reminders to private channels, and tools for managing channel time zones. The messages can be used to remind channel members about calls and events and provide all the necessary links for participation.

## Prerequisites

- go
- golang-migrate (see [Migrations](#migrations))
- docker
- docker-compose

## Installation

### Test

1. (Optional) If you want to test reminder bot on a local mattermost server, you should run `docker-compose up test_mm -d`
   1. If you decide to test the bot like this, you could access mattermost either from browser on <http://localhost:8065> or from desktop client adding a server with the same url
2. Follow [this](https://developers.mattermost.com/integrate/slash-commands/custom/) tutorial to add slash commands to a mattermost server. As a url use POST `http://<host>:8080/mattermost/reminders` (for local server `<host>` is `reminder`)
   1. Probably you will have to add `<host>` to an 'Untrusted internal connections' in mattermost server. To do so, go to `System Console/Environment/Developer` and add your `<host>` to the `Allow untrusted internal connections to:` field.
3. Copy `.env.example` to `.env` and fill in all the required information (see [this section](#env-file) if you got confused about some variables)
4. Run `docker-compose up -d`

### Production

#### Dockerhub

You could use pre-built reminder image from [dockerhub](https://hub.docker.com/repository/docker/andreydrumel/mm-remind-bot/general) as such:

```yaml
# docker-compose.yaml
services:
   reminder:
      image: andreydrumel/mm-remind-bot:1.1.3
      ...
```

```Dockerfile
# Dockerfile
FROM andreydrumel/mm-remind-bot:1.1.3
...
```

Do not forget to provide all the necessary ENV's! They are listed in [Container desription](#container-description) section.

Poller service is not provided in dockerhub. You can either write this service on your own (it's not so hard), use low-code solution (such as n8n) or [build your own image from source](#build-from-source).

#### Build from source

There are Dockerfiles in both `poller` and `reminder` root subdirectories, so you could go in there and run `docker build .`.

Otherwise you could use `docker-compose.prod.yaml` from root directory either as a template or the actual compose-file.

## Usage

`/reminder COMMAND OPTIONS`

Commands:

- `help,h [cron,location,webhook]` - show more descriptive help message about specified command
- `add,create NAME CRON_RULE MESSAGE` - creates new reminder
- `list,ls` - lists all reminders relevant to a current channel
- `delete,del,remove,rm ID...` - deletes a reminders with ID... identifiers
- `timezone,tz LOCATION` - updates channel timezone
- `timezone,tz` - shows current location
- `wh,webhook WEBHOOK` - binds a `WEBHOOK` to the user. After this, the reminder could send messages to any chat the user can
- `chown,own,steal,snatch ID` - steals ownership of the reminder with id `ID`. This changes which webhook the reminder uses to send messages.

### Cron rule

Cron rule is a special string which shows when a remind should be sent. It can be written in three ways:

- `Second Minute Hour DayOfMonth Month DayOfWeek Year`
- `Minute Hour DayOfMonth Month DayOfWeek Year` (Second defaults to 0)
- `Minute Hour DayOfMonth Month DayOfWeek` (Year defaults to *)

There are some notes you want to understand to write appropriate strings:

- Month could be set in two ways: `1-12` or `JAN-DEC`
- Similar for the DayOfWeek: `0-7` or `SUN-SAT` (both `0` and `7` stand for `SUN`)
- `*` - any value (`0 12 * * *` - 12:00 every day every month every year)
- `/` - time period (`*/5 * * * *` - every 5 minute of every day every month every year)
- `,` - list separator (`0 12 10,25 * *` - 12:00 every 10th and 25th day of every month)
- `-` - range (`0 12 * MON-FRI *` - 12:00 every workday)
- `L` - last (`0 12 * 5L *` - 12:00 last friday every month)
- `#` - numbered (`0 12 * TUE#2 *` - 12:00 second tuesday of every month)`

### Location

`LOCATION`: `TZ` identifier (for example `Asia/Novosibirsk`)

You can find all the possible locations [here](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).

### Webhook

A webhook is the access point to Mattermost, which allows the reminder bot to send reminds. The webhook shares access rights with the user who created it. Since users can create private chats, they should use their own webhooks to send messages to these chats.

To use the reminder bot a user should create a webhook and provide it to the bot. Here is the tutorial:

1. Click on the nine squares in the top left corner of your Mattermost client, then click `Integrations`
2. Select `Incoming Webhooks`
3. Click the `Add incoming Webhook` button
4. Choose a title and a channel (you must choose the default one; besides, if everything works fine, it will not be used). Do not check `Lock to this channel` box!
5. Save the webhook and copy the URL you receive on the next screen. The webhook should look like this `http://<mm_host>/hooks/XXXXXX`
6. Go to any chat you have access to and write down `/reminder webhook WEBHOOK` (paste your webhook instead of the caps-locked word). It should be something like `/reminder webhook http://<mm_host>/hooks/XXXXXX`
7. Done! After completing these actions, the reminder bot will send messages wherever you want.

If someone who created a reminder loses access to a chat that the reminder is bound to, you could steal ownership of this reminder using command `/reminder steal ID`. After that, this reminder will send reminds using your webhook (you should specify it first using the tutorial above).

### Examples

Command will create weekly reminder that will be triggered at 12:00 on fridays repeatedly

```text
/reminder add "Weekly Reminder" "0 0 12 * * FRI *" "Another week is coming to an end!"
```

---

Command will create a reminder that triggers at 9:00, 12:00, 15:00 and 18:00 every monday and thursday

```text
/reminder add "Repeatedly reminder" "0 9-18/3 * * MON,THU"
```

## Configuration

Database: MySQL

### .env file

- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `DB_HOST` - DataBase Host - host to access database, defaults to `db` service
- `DB_PORT` - DataBase Port - mysql default is `3306`, but you can change it here
- `DB_NAME` - DataBase Name - default is `reminders`, but if you want to use another name, you should rename it here
- `MM_SC_TOKEN` - MatterMost Slash Command Token - token that you receive after [creating slash command](https://developers.mattermost.com/integrate/slash-commands/custom/)

### Container description

1. `db` - mysql database
   1. `MYSQL_USER`
   2. `MYSQL_PASSWORD`
   3. `MYSQL_DATABASE` == `DB_NAME`
2. `reminder` - server with the main logic which handles a mattermost slash commands and provides useful API endpoints, also runs migrations
   1. `MYSQL_USER`
   2. `MYSQL_PASSWORD`
   3. `DB_HOST`
   4. `DB_PORT`
   5. `DB_NAME`
   6. `MM_SC_TOKEN` - to verify the server attempting to use this command
   7. `DEFAULT_TZ` - Default Time Zone
3. `poller` - simple service that periodically polls the `reminder` container for reminds and sends them to a corresponding mattermost channel using webhook
   1. `POLL_PERIOD` - a time period for `poller` service to poll `reminder` service. Unit suffix are used: `2h45m` stands for 2 hours 45 minutes
4. `test_mm` test profile - container that holds a test local mattermost server

## Migrations

Migrations could be done using [this](https://github.com/golang-migrate/migrate) tool.

To install it, run:

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

(This command will install `migrate` cli program into `$HOME/go/bin` directory, maybe you will have to add this directory to the `$PATH`)

---

To create migration:

```bash
migrate create -ext sql -dir migrations -seq MIGRATION_NAME
```

---

Notice that migrations run automatically when the reminder container starts, so you don't need to do it manually.

To run migrations manually:

```bash
migrate -database 'mysql://MYSQL_USER:MYSQL_PASSWORD@tcp(HOST:PORT)/NAME' -source file://migrations up
```

All the uppercased from the command above you specify in your `.env` file. Using `.env.example` you get:

```bash
migrate -database 'mysql://reminders:XXXXXX@tcp(db:3306)/reminders' -source file://migrations up
```
