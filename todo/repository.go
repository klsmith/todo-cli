package todo

type Repository interface {
	All() []Item
	Count() int
	Find(id ID) (Item, bool)
	Create(description string) (Item, error)
	Update(id ID, description string)
	Complete(id ID) (Item, error)
	UnComplete(id ID) (Item, error)
	Delete(id ID) (Item, error)
	DeleteAll() error
	CleanCompleted() error
}
