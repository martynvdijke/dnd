package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	"villum/middleware"
)

// cmdCategory groups commands in /help and the native command menu.
type cmdCategory string

const (
	catCharacters    cmdCategory = "characters"
	catCampaign      cmdCategory = "campaign"
	catCompendium    cmdCategory = "compendium"
	catNotifications cmdCategory = "notifications"
	catBot           cmdCategory = "bot"
)

var categoryOrder = []cmdCategory{catCharacters, catCampaign, catCompendium, catNotifications, catBot}

var categoryHeadings = map[cmdCategory]string{
	catCharacters:    "🧙 Characters",
	catCampaign:      "🏕 Campaigns",
	catCompendium:    "📚 Compendium",
	catNotifications: "🔔 Notifications",
	catBot:           "🤖 Bot",
}

// command is one entry in the single source of truth for dispatch, /help and
// the native command menu.
type command struct {
	name     string
	usage    string
	desc     string
	category cmdCategory
	aliases  []string
	hidden   bool
	showNav  bool
	run      func(c *cmdContext) botReply
}

var commandRegistry []command

func init() {
	commandRegistry = []command{
		{name: "characters", usage: "/characters", desc: "List your characters", category: catCharacters, showNav: true, run: runCharacters},
		{name: "sheet", usage: "/sheet [name|id]", desc: "Show a character sheet", category: catCharacters, showNav: true, run: runSheet},
		{name: "stats", usage: "/stats [name|id]", desc: "Show character or campaign statistics", category: catCharacters, showNav: true, run: runStats},
		{name: "claim", usage: "/claim [name|id]", desc: "Claim a character as yours", category: catCharacters, run: runClaim},
		{name: "unclaim", usage: "/unclaim", desc: "Release your claimed character", category: catCharacters, run: runUnclaim},
		{name: "hp", usage: "/hp [+N|-N|N]", desc: "Show or adjust HP", category: catCharacters, run: runHP},
		{name: "rest", usage: "/rest short|long", desc: "Take a short or long rest", category: catCharacters, run: runRest},
		{name: "cast", usage: "/cast <spell> [dice]", desc: "Cast a spell", category: catCharacters, run: runCast},
		{name: "create", usage: "/create", desc: "Create a character step by step", category: catCharacters, run: runCreate},
		{name: "overview", usage: "/overview [campaign]", desc: "Show campaign overviews", category: catCampaign, showNav: true, run: runOverview},
		{name: "recap", usage: "/recap [campaign]", desc: "Show recaps for a campaign", category: catCampaign, run: runRecap},
		{name: "lastrecap", usage: "/lastrecap", desc: "Show the most recent recap", category: catCampaign, run: runLastRecap},
		{name: "items", usage: "/items", desc: "List campaign party items", category: catCampaign, run: runItems},
		{name: "quests", usage: "/quests", desc: "List the party's open quests", category: catCampaign, run: runQuests},
		{name: "visits", usage: "/visits", desc: "List party location visits", category: catCampaign, run: runVisits},
		{name: "subscribe", usage: "/subscribe", desc: "Receive recap copies as DMs", category: catNotifications, run: runSubscribe},
		{name: "unsubscribe", usage: "/unsubscribe", desc: "Stop DM recap copies", category: catNotifications, run: runUnsubscribe},
		{name: "status", usage: "/status", desc: "Show your link and claim status", category: catBot, run: runStatus},
		{name: "help", usage: "/help", desc: "Show this command list", category: catBot, showNav: true, run: runHelp},
		{name: "start", usage: "/start <code>", desc: "Link your Villum account", category: catBot, run: runStart},
		{name: "unlink", usage: "/unlink", desc: "Unlink your account", category: catBot, run: runUnlink},
		{name: "cancel", usage: "/cancel", desc: "Cancel the current step-by-step flow", category: catBot, run: runCancel},
		{name: "search", usage: "/search <query>", desc: "Search the compendium", category: catCompendium, showNav: true, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "") }},
		{name: "spell", usage: "/spell <query>", desc: "Search spells", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "spell") }},
		{name: "item", usage: "/item <query>", desc: "Search equipment", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "equipment") }},
		{name: "monster", usage: "/monster <query>", desc: "Search monsters", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "monster") }},
		{name: "race", usage: "/race <query>", desc: "Search races", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "race") }},
		{name: "class", usage: "/class <query>", desc: "Search classes", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "class") }},
		{name: "feat", usage: "/feat <query>", desc: "Search feats", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "feat") }},
		{name: "background", usage: "/background <query>", desc: "Search backgrounds", category: catCompendium, run: func(c *cmdContext) botReply { return runCompendiumSearch(c, "background") }},
		{name: "ask", usage: "/ask <question>", desc: "Ask about the compendium", category: catCompendium, run: func(c *cmdContext) botReply { return runAskQuery(c, strings.Join(c.args, " ")) }},
	}
}

func findCommand(name string) (command, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, cmd := range commandRegistry {
		if cmd.name == name {
			return cmd, true
		}
		for _, alias := range cmd.aliases {
			if alias == name {
				return cmd, true
			}
		}
	}
	return command{}, false
}

// execute runs a command and attaches the navigation keyboard when the
// command opts in and did not provide its own keyboard.
func execute(c *cmdContext, cmd command) botReply {
	r := cmd.run(c)
	if r.Keyboard == nil && cmd.showNav {
		r.Keyboard = navKeyboard()
	}
	return r
}

func helpText() string {
	var b strings.Builder
	b.WriteString("<b>Villum bot commands</b>\n")
	for _, cat := range categoryOrder {
		var lines []string
		for _, cmd := range commandRegistry {
			if cmd.category != cat || cmd.hidden {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s – %s", escapeHTML(cmd.usage), escapeHTML(cmd.desc)))
		}
		if len(lines) == 0 {
			continue
		}
		b.WriteString("\n" + escapeHTML(categoryHeadings[cat]) + "\n")
		b.WriteString(strings.Join(lines, "\n") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func linkingHelpText() string {
	return "🔗 <b>Link your Villum account first</b>\n\n" +
		"1. Open Villum → <b>Settings → Telegram</b>.\n" +
		"2. Generate a link code.\n" +
		"3. Send <code>/start &lt;code&gt;</code> here.\n\n" +
		"Linking unlocks /recap, /characters, /sheet, /stats, /claim and /create."
}

// botCommandList is the native menu payload derived from the registry.
func botCommandList() []tgmodels.BotCommand {
	commands := make([]tgmodels.BotCommand, 0, len(commandRegistry))
	for _, cmd := range commandRegistry {
		if cmd.hidden {
			continue
		}
		desc := cmd.desc
		if len([]rune(desc)) > 256 {
			desc = string([]rune(desc)[:256])
		}
		commands = append(commands, tgmodels.BotCommand{Command: cmd.name, Description: desc})
	}
	return commands
}

// registerBotCommands publishes the registry as the native command menu.
func registerBotCommands(client *tgbot.Bot) {
	if client == nil {
		return
	}
	if _, err := client.SetMyCommands(context.Background(), &tgbot.SetMyCommandsParams{Commands: botCommandList()}); err != nil {
		middleware.LogWarn("telegram", "failed to register command menu", "error", err)
	}
}
