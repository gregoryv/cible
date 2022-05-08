package cible

import "testing"

func TestAreas(t *testing.T) {
	a := Areas{
		&Area{Ident: "i1", Title: "title1"},
		&Area{Ident: "i2", Title: "title2"},
	}
	t.Run("Area", func(t *testing.T) {
		if _, err := a.Area("i1"); err != nil {
			t.Error(err)
		}
		if _, err := a.Area("x"); err == nil {
			t.Error("Area did not return error")
		}
	})
}

func TestTile(t *testing.T) {
	_ = Spaceport()
}
