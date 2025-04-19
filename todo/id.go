package todo

import (
	"strconv"
	"todo/errs"
)

type ID int

type IDs []ID

func IDFromString(s string) (ID, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return ID(-1), errs.Wrap("unable to parse id "+s, err)
	}
	return ID(i), nil
}

func (id ID) DisplayString() string {
	return strconv.Itoa(int(id))
}

func (ids IDs) Len() int {
	return len(ids)
}

func (ids IDs) Swap(i, j int) {
	ids[i], ids[j] = ids[j], ids[i]
}

func (ids IDs) Less(i, j int) bool {
	return ids[i] < ids[j]
}
