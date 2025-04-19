package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"todo/errs"
)

const todoHomeName = ".todo"
const listFileName = "list.json"

var todoPath = ""

func main() {
	defer fmt.Println("")
	args := os.Args[1:]
	usr, err := user.Current()
	errs.MaybePanic("unable to get current user", err)
	userHome := usr.HomeDir
	appHome := filepath.Join(userHome, todoHomeName)
	err = os.MkdirAll(appHome, os.ModePerm)
	errs.MaybePanic("unable to access home directory at "+appHome, err)
	todoPath = filepath.Join(appHome, listFileName)
	initListFile()
	todo := readList()
	err = todo.doCommand(args)
	errs.MaybePanic(fmt.Sprintf("unable to execute %v", args), err)
}

func initListFile() {
	file, err := os.OpenFile(todoPath, os.O_CREATE, os.ModePerm.Perm())
	errs.MaybePanic("unable to open list at "+todoPath, err)
	defer func() {
		err = file.Close()
		errs.MaybePanic("unable to close list file at "+todoPath, err)
	}()
	stat, err := file.Stat()
	errs.MaybePanic("unable to stat list at "+todoPath, err)
	size := stat.Size()
	if size == 0 {
		emptyList := TodoList{LastID: -1, Items: make(TodoItemRefs)}
		err = emptyList.save()
		errs.MaybePanic("unable to generate default to-do list", err)
	}
}

func readList() TodoList {
	bytes, err := os.ReadFile(todoPath)
	errs.MaybePanic("unable to read list at "+todoPath, err)
	var list TodoList
	err = json.Unmarshal(bytes, &list)
	errs.MaybePanic("unable to unmarshal list at "+todoPath, err)
	return list
}

type TodoList struct {
	LastID int          `json:"last_id"`
	Items  TodoItemRefs `json:"items"`
}

type TodoItemRefs = map[int]TodoItem

type TodoItem struct {
	ID          int    `json:"id"`
	Complete    bool   `json:"complete"`
	Description string `json:"description"`
}

func (todo *TodoList) doCommand(mainArgs []string) error {
	var command string
	var commandArgs []string
	if len(mainArgs) == 0 {
		command = "list"
		commandArgs = make([]string, 0)
	} else {
		command = mainArgs[0]
		commandArgs = mainArgs[1:]
	}
	var err error
	switch command {
	case "help":
		doHelpCommand()
	case "list":
		todo.doListCommand()
	case "add":
		err = todo.doAddCommand(commandArgs)
	case "complete":
		err = todo.doCompleteCommand(commandArgs)
	case "remove":
		err = todo.doRemoveCommand(commandArgs)
	case "removeall":
		err = todo.doRemoveAllCommand()
	case "cleanup":
		err = todo.doCleanupCommand()
	}
	if err != nil {
		return err
	}
	return nil
}

func doHelpCommand() {
	fmt.Println("To-Do List CLI")
	fmt.Println("Usage:")
	fmt.Println("\t help \t\t\t You're lookin' at it!")
	fmt.Println("\t list \t\t\t Prints the current state of the list.")
	fmt.Println("\t add \"<description>\" \t Adds a new item to the list.")
	fmt.Println("\t complete <id> \t\t Marks an item as complete.")
	fmt.Println("\t cleanup <id> \t\t Removes all completed items from the list and re-indexes the items.")
	fmt.Println("\t remove <id> \t\t Removes an item from the list.")
	fmt.Println("\t remove all \t\t Removes all items from the list.")
}

func (todo *TodoList) doListCommand() {
	size := todo.size()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
	}
	keys := make([]int, 0, len(todo.Items))
	for k := range todo.Items {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		item := todo.Items[k]
		fmt.Println(item.string())
	}
}

func (todo *TodoList) doAddCommand(args []string) error {
	newDescription := args[0]
	newId := todo.LastID + 1
	todo.LastID = newId
	newItem := TodoItem{
		ID:          newId,
		Description: newDescription,
		Complete:    false,
	}
	todo.Items[newId] = newItem
	fmt.Printf("Added → %s\n", newItem.string())
	return todo.save()
}

func (todo *TodoList) doCompleteCommand(args []string) error {
	item, err := todo.lookupItem(args[0])
	if err != nil {
		return errs.Wrap("unable to complete item "+args[0], err)
	}
	item.Complete = true
	todo.Items[item.ID] = *item
	fmt.Printf("Completed → %s\n", item.string())
	return todo.save()
}

func (todo *TodoList) doRemoveCommand(args []string) error {
	item, err := todo.lookupItem(args[0])
	if err != nil {
		return errs.Wrap("unable to remove item "+args[0], err)
	}
	todo.removeItem(*item)
	return todo.save()
}

func (todo *TodoList) doRemoveAllCommand() error {
	size := todo.size()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
		return nil
	}
	todo.Items = make(TodoItemRefs)
	fmt.Println("Removed all items from to-do list!")
	return todo.save()
}

func (todo *TodoList) removeItem(item TodoItem) {
	delete(todo.Items, item.ID)
	fmt.Printf("Removed → %s\n", item.string())
}

func (todo *TodoList) lookupItem(argId string) (*TodoItem, error) {
	lookupId, err := strconv.Atoi(argId)
	if err != nil {
		return nil, errs.Wrap("unable to parse id "+argId, err)
	}
	item, ok := todo.Items[lookupId]
	if !ok {
		return nil, errs.Wrap("unable to find item "+argId, nil)
	}
	return &item, nil
}

func (todo *TodoList) doCleanupCommand() error {
	size := todo.size()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
		return nil
	}
	fmt.Println("Removing all completed items...")
	for _, item := range todo.Items {
		if item.Complete {
			todo.removeItem(item)
		}
	}
	fmt.Println("Re-indexing list...")
	newSize := todo.size()
	newItems := make(TodoItemRefs, newSize)
	var index = 0
	for _, item := range todo.Items {
		item.ID = index
		newItems[index] = item
		index++
	}
	todo.LastID = newSize
	todo.Items = newItems
	fmt.Println("Cleanup complete!")
	return todo.save()
}

func (todo *TodoList) save() error {
	bytes, err := json.Marshal(todo)
	if err != nil {
		return errs.Wrap("unable save changes", err)
	}
	err = os.WriteFile(todoPath, bytes, os.ModePerm.Perm())
	if err != nil {
		return errs.Wrap("unable save changes "+todoPath, err)
	}
	return nil
}

func (todo *TodoList) size() int {
	return len(todo.Items)
}

func (item *TodoItem) string() string {
	var complete string
	if item.Complete {
		complete = "✅"
	} else {
		complete = "❌"
	}
	return fmt.Sprintf("%3d: %-2s %s", item.ID, complete, item.Description)
}
