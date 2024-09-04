## Installation

1. Copy `.env.example` to `.env` and fill in all the database information
2. Type `go run .` to run a server
3. Follow [this](https://developers.mattermost.com/integrate/slash-commands/custom/) tutorial to add slash commands, as a url use POST `http://<host>:8080/mattermost/reminders`

## Usage

Bot has several commands:
- `create [NAME] [RULE]` - creates new reminder with the specified rule
- `list` - shows the reminders table with their names, rules and channels
