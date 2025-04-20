package main

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
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
	cmd := buildCmd(repo)
	err = cmd.Run(context.Background(), os.Args)
	errs.MaybePanic(fmt.Sprintf("unable to execute %v", args), err)
	//err = doCommand(args, repo)
	errs.MaybePanic(fmt.Sprintf("unable to execute %v", args), err)
}

func buildCmd(repo todo.Repository) *cli.Command {
	return &cli.Command{
		Name:  "todo",
		Usage: "maintain a to-do checklist from your terminal",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			doListCommand(repo)
			return nil
		},
		Commands: []*cli.Command{
			{Usage: "calling without arguments defaults to the list command"},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "print the current state of the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					doListCommand(repo)
					return nil
				},
			},
			{
				Name:      "add",
				Aliases:   []string{"a"},
				Usage:     "add a new item to the list",
				ArgsUsage: "<description>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "description"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doAddCommand(cmd, repo)
				},
			},
			{
				Name:      "edit",
				Aliases:   []string{"e"},
				Usage:     "open the item description for editing",
				ArgsUsage: "<id>",
				Arguments: []cli.Argument{
					&cli.IntArg{Name: "id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doEditCommand(cmd, repo)
				},
			},
			{
				Name:      "complete",
				Aliases:   []string{"c"},
				Usage:     "mark an item as complete",
				ArgsUsage: "<id>",
				Arguments: []cli.Argument{
					&cli.IntArg{Name: "id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doCompleteCommand(cmd, repo)
				},
			},
			{
				Name:      "incomplete",
				Aliases:   []string{"i"},
				Usage:     "mark an item as incomplete",
				ArgsUsage: "<id>",
				Arguments: []cli.Argument{
					&cli.IntArg{Name: "id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doIncompleteCommand(cmd, repo)
				},
			},
			{
				Name:      "remove",
				Aliases:   []string{"r"},
				Usage:     "remove an item from the list",
				ArgsUsage: "<id>",
				Arguments: []cli.Argument{
					&cli.IntArg{Name: "id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doRemoveCommand(cmd, repo)
				},
			},
			{
				Name:    "purge",
				Aliases: []string{"p"},
				Usage:   "remove all items from the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doPurgeCommand(repo)
				},
			},
			{
				Name:    "cleanup",
				Aliases: []string{"cu"},
				Usage:   "remove all completed items from the list",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return doCleanupCommand(repo)
				},
			},
		},
	}
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

func doAddCommand(cmd *cli.Command, repo todo.Repository) error {
	newDescription := cmd.Args().First()
	item, err := repo.Create(newDescription)
	if err != nil {
		return err
	}
	fmt.Printf("Added → %s\n", item.DisplayString())
	return nil
}

func doEditCommand(cmd *cli.Command, repo todo.Repository) error {
	idToParse := cmd.Args().First()
	id, err := todo.IDFromString(idToParse)
	if err != nil {
		return errs.Wrap("unable to edit item "+idToParse, err)
	}
	item, found := repo.Find(id)
	if !found {
		return errs.New("unable to edit item " + id.DisplayString())
	}
	newDescription, err := pterm.InteractiveTextInputPrinter{
		DefaultText:  "Edit Description",
		Delimiter:    ": ",
		TextStyle:    &pterm.ThemeDefault.DefaultText,
		DefaultValue: item.Description,
	}.Show()
	if err != nil {
		return errs.Wrap("error while editing item "+id.DisplayString(), err)
	}
	updatedItem, err := repo.Update(id, newDescription)
	if err != nil {
		return err
	}
	fmt.Printf("Updated → %s\n", updatedItem.DisplayString())
	return nil
}

func doCompleteCommand(cmd *cli.Command, repo todo.Repository) error {
	id, err := todo.IDFromString(cmd.Args().First())
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

func doIncompleteCommand(cmd *cli.Command, repo todo.Repository) error {
	id, err := todo.IDFromString(cmd.Args().First())
	if err != nil {
		return errs.Wrap("unable to incomplete item", err)
	}
	item, err := repo.Incomplete(id)
	if err != nil {
		return errs.Wrap("unable to incomplete item "+id.DisplayString(), err)
	}
	fmt.Printf("Incompleted → %s\n", item.DisplayString())
	return nil
}

func doRemoveCommand(cmd *cli.Command, repo todo.Repository) error {
	id, err := todo.IDFromString(cmd.Args().First())
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
