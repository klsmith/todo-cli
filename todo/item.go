package todo

import "fmt"

type Item struct {
	ID          ID
	Complete    bool
	Description string
}

func (item *Item) DisplayString() string {
	var complete string
	if item.Complete {
		complete = "✅"
	} else {
		complete = "❌"
	}
	return fmt.Sprintf("%3d: %-2s %s", item.ID, complete, item.Description)
}
