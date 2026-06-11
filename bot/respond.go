package bot

import (
	"log"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// reply sends an immediate text reply to a slash command.
func reply(e *events.ApplicationCommandInteractionCreate, content string) {
	if err := e.CreateMessage(discord.NewMessageCreate().WithContent(content)); err != nil {
		log.Printf("failed responding to interaction: %v", err)
	}
}

// followup sends a text follow-up after a deferred slash-command response.
func followup(e *events.ApplicationCommandInteractionCreate, content string) {
	followupComplex(e, discord.NewMessageCreate().WithContent(content))
}

// followupComplex sends a rich follow-up (embeds/components) after a deferred
// slash-command response.
func followupComplex(e *events.ApplicationCommandInteractionCreate, mc discord.MessageCreate) {
	if _, err := e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), mc); err != nil {
		log.Printf("failed sending follow-up: %v", err)
	}
}

// editComponent edits the deferred response of a component interaction,
// replacing the content and clearing the embeds/buttons (used after a search
// selection).
func editComponent(e *events.ComponentInteractionCreate, content string) {
	if _, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(),
		discord.NewMessageUpdate().WithContent(content).ClearEmbeds().ClearComponents()); err != nil {
		log.Printf("failed updating interaction response: %v", err)
	}
}
