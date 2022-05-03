package cible

import "testing"

func TestAreas(t *testing.T) {
	t.Run("Area unknown", func(t *testing.T) {
		a := make(Areas, 0)
		if _, err := a.Area("unknown"); err == nil {
			t.Fail()
		}
	})
}
