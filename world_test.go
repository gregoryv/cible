package cible

import (
	"testing"
)

func TestWorld(t *testing.T) {
	w := NewWorld()
	a := NewArea()
	a.AddTile(NewTile())
	a.AddTile(NewTile())
	if err := a.SetLinks([]Link{
		{0, 1, East},
	}); err != nil {
		t.Fatal(err)
	}
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

	t.Run("Move", func(t *testing.T) {
		if _, err := w.Move(East, From{0, 0}); err != nil {
			t.Error(err)
		}
		if _, err := w.Move(West, From{0, 1}); err != nil {
			t.Error(err)
		}
		t.Run("wrong direction", func(t *testing.T) {
			if _, err := w.Move(West, From{0, 0}); err == nil {
				t.Error(err)
			}
		})
		t.Run("from unknown location", func(t *testing.T) {
			if _, err := w.Move(West, From{1, 1}); err == nil {
				t.Error(err)
			}
		})
	})

}
