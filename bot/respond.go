package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// respond sends an immediate text reply to an interaction.
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
	if err != nil {
		log.Printf("failed responding to interaction: %v", err)
	}
}

// deferResponse acknowledges an interaction so the bot can follow up after a
// slow operation (search / link resolution).
func deferResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("failed deferring interaction: %v", err)
	}
}

// followup sends a text follow-up after a deferred response.
func followup(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	followupComplex(s, i, &discordgo.WebhookParams{Content: content})
}

// followupComplex sends a rich follow-up (embeds/components) after a deferred
// response.
func followupComplex(s *discordgo.Session, i *discordgo.InteractionCreate, params *discordgo.WebhookParams) {
	if _, err := s.FollowupMessageCreate(i.Interaction, true, params); err != nil {
		log.Printf("failed sending follow-up: %v", err)
	}
}

// updateMessage edits the message a component belongs to, replacing its content
// and clearing its components (used after a search selection).
func updateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		log.Printf("failed updating message: %v", err)
	}
}
