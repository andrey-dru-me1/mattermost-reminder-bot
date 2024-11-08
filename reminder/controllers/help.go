package controllers

func usage() string {
	return "Usage: `/reminder COMMAND OPTIONS`\n" +
		"Commands:\n\n" +

		"- `help,h [cron,location,webhook]` - show more descriptive help message about specified command\n" +
		"- `add,create NAME CRON_RULE MESSAGE` - creates new reminder\n" +
		"- `list,ls` - lists all reminders\n" +
		"- `delete,del,remove,rm ID...` - deletes a reminders with ID... identifiers\n" +
		"- `timezone,tz LOCATION` - updates channel timezone\n" +
		"- `timezone,tz` - shows current location\n" +
		"- `wh,webhook WEBHOOK` - binds a `WEBHOOK` to the user. After this, the reminder could send messages to any chat the user can\n" +
		"- `chown,own,steal,snatch ID` - steals ownership of the reminder with id `ID`. This changes which webhook the reminder uses to send messages.\n"
}

func helpCronRule() string {
	return "Cron rule is a special string which shows when a remind should be sent. It can be written in three ways:\n\n" +

		"- `Second Minute Hour DayOfMonth Month DayOfWeek Year`\n" +
		"- `Minute Hour DayOfMonth Month DayOfWeek Year` (Second defaults to `0`)\n" +
		"- `Minute Hour DayOfMonth Month DayOfWeek` (Year defaults to `*`)\n\n" +

		"There are some notes you want to understand to write appropriate strings:\n\n" +

		"- Month could be set in two ways: `1-12` or `JAN-DEC`\n" +
		"- Similar for the DayOfWeek: `0-7` or `SUN-SAT` (both `0` and `7` stand for `SUN`)\n" +
		"- `*` - any value (`0 12 * * *` - 12:00 every day every month every year)\n" +
		"- `/` - time period (`*/5 * * * *` - every 5 minute of every day every month every year)\n" +
		"- `,` - list separator (`0 12 10,25 * *` - 12:00 every 10th and 25th day of every month)\n" +
		"- `-` - range (`0 12 * MON-FRI *` - 12:00 every workday)\n" +
		"- `L` - last (`0 12 * 5L *` - 12:00 last friday every month)\n" +
		"- `#` - numbered (`0 12 * TUE#2 *` - 12:00 second tuesday of every month)`\n"
}

func helpLocation() string {
	return "`LOCATION`: `TZ` identifier (for example `Asia/Novosibirsk`)\n" +
		"You can find all the possible locations [here](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)."
}

func helpWebhook() string {
	return "To use the reminder bot, you need to create a webhook and provide it to the bot. Here is the tutorial:\n\n" +

		"1. Click the nine squares in the top left of Mattermost.\n" +
		"2. Go to `Integrations > Incoming Webhooks > Add Incoming Webhook`, fill in title and channel and save the webhook.\n" +
		"3. Copy the URL you receive, then go to any channel and enter: `/reminder webhook http://<host>/hooks/XXXXXX` (paste the copied URL as is).\n\n" +

		"For more details, visit [this link](https://github.com/andrey-dru-me1/mattermost-reminder-bot/tree/v1.1.2?tab=readme-ov-file#webhook)."
}

func help(tokens []string) string {
	if len(tokens) <= 1 {
		return usage()
	}
	switch tokens[1] {
	case "webhook", "wh":
		return helpWebhook()
	case "location", "loc", "timezone", "tz":
		return helpLocation()
	case "cron", "cronrule", "cronexpr":
		return helpCronRule()
	default:
		return usage()
	}
}
