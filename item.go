package cible

import "errors"

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

func (me Items) FindByName(n Name) (*Item, error) {
	for _, item := range me {
		if item.Name == item.Name {
			return item, nil
		}
	}
	return nil, ErrItemNotFound
}

var ErrItemNotFound = errors.New("item not found")

type Item struct {
	Name
	Count uint

	Location // if it's not in a persons inventory
}
