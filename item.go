package cible

type Items []*Item

func (me *Items) At(loc Location) Items {
	res := make(Items, 0)
	for _, item := range *me {
		if item.Location.Equal(loc) {
			res = append(res, item)
		}
	}
	return res
}

type Item struct {
	Name
	Count uint

	Location // if it's not in a persons inventory
}
