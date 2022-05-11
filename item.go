package cible

type Items []Item

func (me *Items) At(p Position) Items {
	res := make(Items, 0)
	for _, item := range *me {
		if item.Position.Equal(p) {
			res = append(res, item)
		}
	}
	return res
}

type Item struct {
	Name
	Count uint

	Position // if it's not in a persons inventory
}
