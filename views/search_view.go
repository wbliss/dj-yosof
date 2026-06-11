// Package views builds Discord message components, porting djyosof/views.
package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"

	"github.com/GusPrice/dj-yosof/audio"
)

// ComponentPrefix is the custom-id prefix for search-result selection buttons.
const ComponentPrefix = "select"

// SearchComponents builds a row of numbered buttons (1..len(tracks), max 5),
// one per search result. The custom id encodes the session key and the result
// index so the handler can look up the chosen track. Ports SearchView /
// SearchResultButton.
func SearchComponents(sessionKey string, count int) []discord.LayoutComponent {
	if count > 5 {
		count = 5
	}
	buttons := make([]discord.InteractiveComponent, 0, count)
	for i := 0; i < count; i++ {
		buttons = append(buttons, discord.NewPrimaryButton(
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%s:%s:%d", ComponentPrefix, sessionKey, i),
		))
	}
	return []discord.LayoutComponent{discord.NewActionRow(buttons...)}
}

// SearchEmbed renders the list of search results as an embed.
func SearchEmbed(tracks []audio.PlayableAudio) discord.Embed {
	var b strings.Builder
	for i, t := range tracks {
		fmt.Fprintf(&b, "%d. %s\n", i+1, t.DisplayName())
	}
	return discord.NewEmbed().
		WithTitle("Search Results").
		WithDescription(b.String())
}

// ParseComponentID splits a selection custom id into its session key and result
// index. ok is false if the id is not a selection id.
func ParseComponentID(customID string) (sessionKey string, index int, ok bool) {
	parts := strings.Split(customID, ":")
	if len(parts) != 3 || parts[0] != ComponentPrefix {
		return "", 0, false
	}
	index, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, false
	}
	return parts[1], index, true
}
