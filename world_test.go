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

}
