package playlist

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Rules describes how a smart playlist resolves its tracks.
// it is stored as JSON in a playlist's `#GONIC-RULES` m3u attribute.
// all fields are optional; they are ANDed together when present.
type Rules struct {
	Genre    string `json:"genre,omitempty"`
	FromYear int    `json:"fromYear,omitempty"`
	ToYear   int    `json:"toYear,omitempty"`
	// MinRating is the minimum 1-5 star rating from any user
	MinRating int `json:"minRating,omitempty"`
	// Starred limits to tracks starred by any user
	Starred bool `json:"starred,omitempty"`
	// MinPlays is the minimum play count from any user
	MinPlays int `json:"minPlays,omitempty"`
	// Sort is one of: random (default), rating, plays, year, title
	Sort  string `json:"sort,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

func ParseRules(in string) (*Rules, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return nil, nil
	}
	rules := &Rules{}
	if err := json.Unmarshal([]byte(in), rules); err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	switch rules.Sort {
	case "", "random", "rating", "plays", "year", "title":
	default:
		return nil, fmt.Errorf("unknown sort %q", rules.Sort)
	}
	if rules.MinRating < 0 || rules.MinRating > 5 {
		return nil, fmt.Errorf("minRating must be 0-5, got %d", rules.MinRating)
	}
	return rules, nil
}

func (r *Rules) String() string {
	if r == nil {
		return ""
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(buf)
}
