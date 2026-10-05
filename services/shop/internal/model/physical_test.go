package model

import "testing"

func TestIsPhysicalItem(t *testing.T) {
	digital := []string{"game_skin", "Game_Skin", "profile_theme", "skin", "theme", "card", "карточка"}
	for _, c := range digital {
		if IsPhysicalItem(c) {
			t.Fatalf("%q should be digital", c)
		}
	}
	physical := []string{"merch", "MERCH", "мерч", "брелок", "unknown", ""}
	for _, c := range physical {
		if !IsPhysicalItem(c) {
			t.Fatalf("%q should be physical", c)
		}
	}
}
