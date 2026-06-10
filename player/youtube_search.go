package player

import (
	"encoding/json"
	"fmt"
	"strings"
)

// videoRenderer is the subset of YouTube's videoRenderer object we need.
type videoRenderer struct {
	VideoID string `json:"videoId"`
	Title   struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"title"`
	Thumbnail struct {
		Thumbnails []struct {
			URL string `json:"url"`
		} `json:"thumbnails"`
	} `json:"thumbnail"`
	LengthText struct {
		SimpleText string `json:"simpleText"`
	} `json:"lengthText"`
}

// ytInitialData captures the path to search results inside the ytInitialData
// blob embedded in the results page.
type ytInitialData struct {
	Contents struct {
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents []struct {
						ItemSectionRenderer struct {
							Contents []struct {
								VideoRenderer *videoRenderer `json:"videoRenderer"`
							} `json:"contents"`
						} `json:"itemSectionRenderer"`
					} `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
	} `json:"contents"`
}

// parseSearchResults extracts the videoRenderer entries from a results page.
func parseSearchResults(html string) ([]videoRenderer, error) {
	raw, err := extractJSONObject(html, "ytInitialData")
	if err != nil {
		return nil, fmt.Errorf("youtube: locating ytInitialData: %w", err)
	}

	var data ytInitialData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("youtube: decoding ytInitialData: %w", err)
	}

	var out []videoRenderer
	for _, section := range data.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents {
		for _, item := range section.ItemSectionRenderer.Contents {
			if item.VideoRenderer != nil {
				out = append(out, *item.VideoRenderer)
			}
		}
	}
	return out, nil
}

// extractJSONObject finds `marker` in s and returns the brace-balanced JSON
// object that follows the next '{'. It respects string literals and escapes so
// braces inside strings don't confuse the matcher.
func extractJSONObject(s, marker string) (string, error) {
	idx := strings.Index(s, marker)
	if idx < 0 {
		return "", fmt.Errorf("marker %q not found", marker)
	}
	start := strings.IndexByte(s[idx:], '{')
	if start < 0 {
		return "", fmt.Errorf("no object after marker %q", marker)
	}
	start += idx

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("unbalanced object after marker %q", marker)
}
