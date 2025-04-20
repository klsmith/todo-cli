package todo

import (
	"encoding/json"
	"os"
	"sort"
	"todo/errs"
)

type ListJson struct {
	LastID ID          `json:"last_id"`
	Items  ItemJsonMap `json:"items"`
}

type ItemJsonMap = map[ID]ItemJson

type ItemJson struct {
	ID          ID     `json:"id"`
	Complete    bool   `json:"complete"`
	Description string `json:"description"`
}

type JsonRepository struct {
	jsonFilePath string
	list         *ListJson
}

func NewJsonRepository(filepath string) *JsonRepository {
	newJsonRepository := &JsonRepository{jsonFilePath: filepath, list: nil}
	newJsonRepository.load()
	return newJsonRepository
}

func itemToJson(item Item) ItemJson {
	return ItemJson{
		ID:          item.ID,
		Complete:    item.Complete,
		Description: item.Description,
	}
}

func jsonToItem(json ItemJson) Item {
	return Item{
		ID:          json.ID,
		Complete:    json.Complete,
		Description: json.Description,
	}
}

func (r *JsonRepository) load() {
	if r.isJsonFileInitialized() {
		bytes, err := os.ReadFile(r.jsonFilePath)
		errs.MaybePanic("unable to read list at "+r.jsonFilePath, err)
		var list ListJson
		err = json.Unmarshal(bytes, &list)
		errs.MaybePanic("unable to unmarshal list at "+r.jsonFilePath, err)
		r.list = &list
	} else {
		r.list = &ListJson{LastID: -1, Items: make(ItemJsonMap)}
		err := r.save()
		errs.MaybePanic("unable to generate default to-do list", err)
	}
}

func (r *JsonRepository) isJsonFileInitialized() bool {
	file, err := os.OpenFile(r.jsonFilePath, os.O_CREATE, os.ModePerm.Perm())
	errs.MaybePanic("unable to open list at "+r.jsonFilePath, err)
	defer func() {
		err = file.Close()
		errs.MaybePanic("unable to close list file at "+r.jsonFilePath, err)
	}()
	stat, err := file.Stat()
	if err != nil && os.IsNotExist(err) {
		return false
	}
	errs.MaybePanic("unable to stat list at "+r.jsonFilePath, err)
	size := stat.Size()
	return size != 0
}

func (r *JsonRepository) save() error {
	bytes, err := json.Marshal(r.list)
	if err != nil {
		return errs.Wrap("unable save changes", err)
	}
	err = os.WriteFile(r.jsonFilePath, bytes, os.ModePerm.Perm())
	if err != nil {
		return errs.Wrap("unable save changes "+r.jsonFilePath, err)
	}
	return nil
}

func (r *JsonRepository) Count() int {
	return len(r.list.Items)
}

func (r *JsonRepository) All() []Item {
	keys := IDs(make([]ID, 0, len(r.list.Items)))
	for k := range r.list.Items {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	var results []Item
	for _, k := range keys {
		results = append(results, jsonToItem(r.list.Items[k]))
	}
	return results
}

func (r *JsonRepository) Find(id ID) (Item, bool) {
	itemJson, found := r.findJson(id)
	return jsonToItem(itemJson), found
}

func (r *JsonRepository) findJson(id ID) (ItemJson, bool) {
	itemJson, found := r.list.Items[id]
	return itemJson, found
}

func (r *JsonRepository) Create(description string) (Item, error) {
	newID := r.list.LastID + 1
	r.list.LastID = newID
	item := Item{
		ID:          newID,
		Description: description,
		Complete:    false,
	}
	r.list.Items[item.ID] = itemToJson(item)
	err := r.save()
	if err != nil {
		return item, errs.Wrap("unable to create item "+description, err)
	}
	return item, nil
}

func (r *JsonRepository) Update(id ID, description string) (Item, error) {
	itemJson, found := r.findJson(id)
	if !found {
		return Item{}, errs.New("unable to update item " + id.DisplayString())
	}
	itemJson.Description = description
	r.list.Items[id] = itemJson
	err := r.save()
	if err != nil {
		return Item{}, errs.Wrap("unable to update item "+id.DisplayString(), err)
	}
	return jsonToItem(itemJson), nil
}

func (r *JsonRepository) Complete(id ID) (Item, error) {
	item, found := r.Find(id)
	if !found {
		return Item{}, errs.New("unable to complete item " + id.DisplayString())
	}
	item.Complete = true
	r.list.Items[item.ID] = itemToJson(item)
	err := r.save()
	if err != nil {
		return item, errs.Wrap("unable to complete item "+id.DisplayString(), err)
	}
	return item, nil
}

func (r *JsonRepository) UnComplete(id ID) (Item, error) {
	item, found := r.Find(id)
	if !found {
		return Item{}, errs.New("unable to un-complete item " + id.DisplayString())
	}
	item.Complete = false
	r.list.Items[item.ID] = itemToJson(item)
	err := r.save()
	if err != nil {
		return item, errs.Wrap("unable to un-complete item "+id.DisplayString(), err)
	}
	return item, nil
}

func (r *JsonRepository) Delete(id ID) (Item, error) {
	item, found := r.Find(id)
	if !found {
		return Item{}, errs.New("unable to delete item " + id.DisplayString())
	}
	delete(r.list.Items, id)
	err := r.save()
	if err != nil {
		return item, errs.Wrap("unable to delete item "+id.DisplayString(), err)
	}
	return item, nil
}

func (r *JsonRepository) DeleteAll() error {
	r.list.Items = make(ItemJsonMap)
	err := r.save()
	if err != nil {
		return errs.Wrap("unable to delete all items", err)
	}
	return nil
}

func (r *JsonRepository) CleanCompleted() error {
	for _, item := range r.list.Items {
		if item.Complete {
			delete(r.list.Items, item.ID)
		}
	}
	err := r.save()
	if err != nil {
		return errs.Wrap("unable to clean complete items", err)
	}
	return nil
}
