package cible

import "testing"

func TestWorld(t *testing.T) {
	_ = NewWorld()
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
		okLinks := []Link{{0, 1, East}}
		if err := a.SetLinks(okLinks); err != nil {
			t.Error(err)
		}

		badLinks := []Link{{0, 3, South}}
		if err := a.SetLinks(badLinks); err == nil {
			t.Error("area does not have tile with id 3, should fail")
		}

	})

}
