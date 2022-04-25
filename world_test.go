package cible

import "testing"

func TestWorld(t *testing.T) {
	w := NewWorld()
	a := NewArea()
	a.AddTile(NewTile())
	a.AddTile(NewTile())
	a.SetLinks([]Link{
		{0, 1, East},
	})
	w.AddArea(a)
	t.Run("Location", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			if _, err := w.Location(At{0, 0}); err != nil {
				t.Error(err)
			}
		})
		t.Run("", func(t *testing.T) {
			if _, err := w.Location(At{10, 0}); err == nil {
				t.Error(err)
			}
		})
		t.Run("", func(t *testing.T) {
			if _, err := w.Location(At{0, 10}); err == nil {
				t.Error(err)
			}
		})
	})
}

func TestArea(t *testing.T) {
	t.Run("defaults to empty", func(t *testing.T) {
		if a := NewArea(); a.Size() != 0 {
			t.Error("not empty, got:", a.Size())
		}
	})

	t.Run("AddTile", func(t *testing.T) {
		a := NewArea()
		a.AddTile(NewTile())
		a.AddTile(NewTile())
		if a.Size() != 2 {
			t.Fail()
		}
	})

	t.Run("SetLinks", func(t *testing.T) {
		a := NewArea()
		a.AddTile(NewTile())
		a.AddTile(NewTile())
		a.AddTile(NewTile())
		okLinks := []Link{
			{0, 1, East},
			{1, 2, South},
		}
		if err := a.SetLinks(okLinks); err != nil {
			t.Error(err)
		}

		badCases := map[string][]Link{
			"missing A":              []Link{{3, 0, South}},
			"missing B":              []Link{{0, 3, South}},
			"missing link to tile 2": []Link{{0, 1, East}},
		}
		for name, links := range badCases {
			t.Run(name, func(t *testing.T) {
				if err := a.SetLinks(links); err == nil {
					t.Fail()
				}
			})
		}
	})
}
