package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"todo/errs"
	"todo/todo"
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
	repo := todo.NewJsonRepository(todoPath)
	err = doCommand(args, repo)
	errs.MaybePanic(fmt.Sprintf("unable to execute %v", args), err)
}

func doCommand(mainArgs []string, repo todo.Repository) error {
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
		doListCommand(repo)
	case "add":
		err = doAddCommand(commandArgs, repo)
	case "complete":
		err = doCompleteCommand(commandArgs, repo)
	case "uncomplete":
		err = doUnCompleteCommand(commandArgs, repo)
	case "remove":
		err = doRemoveCommand(commandArgs, repo)
	case "purge":
		err = doPurgeCommand(repo)
	case "cleanup":
		err = doCleanupCommand(repo)
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

func doListCommand(repo todo.Repository) {
	size := repo.Count()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
	}
	items := repo.All()
	for _, item := range items {
		fmt.Println(item.DisplayString())
	}
}

func doAddCommand(args []string, repo todo.Repository) error {
	newDescription := args[0]
	item, err := repo.Create(newDescription)
	if err != nil {
		return err
	}
	fmt.Printf("Added → %s\n", item.DisplayString())
	return nil
}

func doCompleteCommand(args []string, repo todo.Repository) error {
	id, err := todo.IDFromString(args[0])
	if err != nil {
		return errs.Wrap("unable to complete item", err)
	}
	item, err := repo.Complete(id)
	if err != nil {
		return errs.Wrap("unable to complete item "+id.DisplayString(), err)
	}
	fmt.Printf("Completed → %s\n", item.DisplayString())
	return nil
}

func doUnCompleteCommand(args []string, repo todo.Repository) error {
	id, err := todo.IDFromString(args[0])
	if err != nil {
		return errs.Wrap("unable to un-complete item", err)
	}
	item, err := repo.UnComplete(id)
	if err != nil {
		return errs.Wrap("unable to un-complete item "+id.DisplayString(), err)
	}
	fmt.Printf("Un-Completed → %s\n", item.DisplayString())
	return nil
}

func doRemoveCommand(args []string, repo todo.Repository) error {
	id, err := todo.IDFromString(args[0])
	if err != nil {
		return errs.Wrap("unable to remove item", err)
	}
	item, err := repo.Delete(id)
	if err != nil {
		return errs.Wrap("unable to remove item "+id.DisplayString(), err)
	}
	fmt.Printf("Removed → %s\n", item.DisplayString())
	return nil
}

func doPurgeCommand(repo todo.Repository) error {
	size := repo.Count()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
	}
	err := repo.DeleteAll()
	if err != nil {
		return errs.Wrap("unable to remove all items", err)
	}
	fmt.Println("Removed all items from to-do list!")
	return nil
}

func doCleanupCommand(repo todo.Repository) error {
	size := repo.Count()
	if size == 0 {
		fmt.Println("Your to-do list is empty!")
	}
	fmt.Println("Removing all completed items...")
	err := repo.CleanCompleted()
	if err != nil {
		return errs.Wrap("unable to remove all completed items", err)
	}
	fmt.Println("Cleanup complete!")
	return nil
}
