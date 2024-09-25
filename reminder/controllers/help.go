package controllers

const usage = "Usage: `/reminder COMMAND OPTIONS`\n" +
	"Commands:\n" +
	"- `help,h [cron,location,webhook]` - show more descriptive help message about specified command\n" +
	"- `add,create NAME CRON_RULE MESSAGE` - creates new reminder\n" +
	"- `list,ls` - lists all reminders\n" +
	"- `delete,del,remove,rm ID...` - deletes a reminders with ID... identifiers\n" +
	"- `timezone,tz LOCATION` - updates channel timezone\n" +
	"- `timezone,tz` - shows current location\n" +
	"- `wh,webhook WEBHOOK` - binds a `WEBHOOK` to the user. After this, the reminder could send messages to any chat the user can\n" +
	"- `chown,own,steal,snatch ID` - steals ownership of the reminder with id `ID`. This changes which webhook the reminder uses to send messages.\n"

func helpCronRule() string {
	return "Cron rule is a special string which shows when a remind should be sent. It can be written in three ways:\n\n" +
		"- `Second Minute Hour DayOfMonth Month DayOfWeek Year`\n" +
		"- `Minute Hour DayOfMonth Month DayOfWeek Year` (Second defaults to 0)\n" +
		"- `Minute Hour DayOfMonth Month DayOfWeek` (Year defaults to *)\n\n" +
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
	return "A webhook is the access point to Mattermost, which allows the reminder bot to send reminds. The webhook shares access rights with the user who created it. Since users can create private chats, they should use their own webhooks to send messages to these chats.\n" +
		"To use the reminder bot a user should create a webhook and provide it to the bot. Here is the tutorial:\n\n" +
		"1. Click on the nine squares in the top left corner of your Mattermost client, then click `Integrations`\n" +
		"2. Select `Incoming Webhooks`\n" +
		"3. Click the `Add incoming Webhook` button\n" +
		"4. Choose a title and a channel (you must choose the default one; besides, if everything works fine, it will not be used). Do not check `Lock to this channel` box!\n" +
		"5. Save the webhook and copy the URL you receive on the next screen. The webhook should look like this `http://localhost:8065/hooks/XXXXXX`\n" +
		"6. Go to any chat you have access to and write down `/reminder webhook WEBHOOK` (paste your webhook instead of the caps-locked word). It should be something like `/reminder webhook http://localhost:8065/hooks/XXXXXX`\n" +
		"7. Done! After completing these actions, the reminder bot will send messages wherever you want.\n\n" +
		"If someone who created a reminder loses access to a chat that the reminder is bound to, you could steal ownership of this reminder using command `/reminder steal ID`. After that, this reminder will send reminds using your webhook (you should specify it first using the tutorial above).\n"
}

func help(tokens []string) string {
	if len(tokens) <= 1 {
		return usage
	}
	switch tokens[1] {
	case "webhook", "wh":
		return helpWebhook()
	case "location", "loc", "timezone", "tz":
		return helpLocation()
	case "cron", "cronrule", "cronexpr":
		return helpCronRule()
	default:
		return usage
	}
}
