package model

import "strings"

// IsPhysicalItem reports whether a catalog category is physical (merch).
// Digital whitelist: game_skin, profile_theme, skin, theme, card (+ RU aliases).
// Unknown categories default to physical (safer: block refund after fulfill).
// Live shop DB today: merch | game_skin (English).
func IsPhysicalItem(category string) bool {
	c := strings.ToLower(strings.TrimSpace(category))
	switch c {
	case "game_skin", "profile_theme", "skin", "theme", "card", "карточка":
		return false
	default:
		return true
	}
}
