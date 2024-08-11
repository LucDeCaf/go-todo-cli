package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"
)

type Todo struct {
	Description string
	Completed   bool
	CreatedAt   time.Time
}

// TODO: Use slice instead for 'free' ordered data
var todos map[int]*Todo = make(map[int]*Todo)
var nextId int = 1

func main() {
	const FilePath = "todo_data.csv"

	args := os.Args[1:]
	argCount := len(args)
	if argCount < 1 {
		os.Exit(1)
		return
	}

	// Read existing todos from file
	if err := ReadFromCSV(FilePath); err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdin, 0, 0, 4, ' ', 0)

	action := args[0]
	switch action {
	case "list":
		fmt.Fprint(w, "ID\tDescription\tCompleted\tCreated At\n")

		for id, t := range todos {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", id, t.Description, t.Completed, t.CreatedAt)
		}

		if err := w.Flush(); err != nil {
			log.Fatal(err)
		}
	case "add":
		if argCount < 2 {
			log.Fatal("Err: Missing argument 'description' for action 'add'")
		}
		if argCount > 2 {
			log.Fatal("Err: Too many arguments for action 'add'")
		}

		t := NewTodo(args[1])
		todos[nextId] = &t

		fmt.Printf("Added todo '%v' with id '%v'\n", args[1], nextId)
	case "check":
		if argCount < 2 {
			log.Fatal("Err: Missing argument 'id' for action 'check'")
		}
		if argCount > 2 {
			log.Fatal("Err: Too many arguments for action 'check'")
		}

		id, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatal("Err: " + err.Error())
		}

		if todos[id].Completed {
			fmt.Println("Already completed.")
			break
		}
		
		todos[id].Completed = true
		fmt.Printf("Marked '%v' as completed\n", todos[id].Description)
	default:
		log.Fatal(fmt.Sprintf("Err: Invalid action '%v'\n", action))
	}

	// Write updated todos from program memory
	if err := WriteToCSV(FilePath); err != nil {
		log.Fatal(err)
	}

}

func NewTodo(desc string) Todo {
	t := Todo{
		Description: desc,
		Completed:   false,
		CreatedAt:   time.Now(),
	}
	return t
}

func NewTodoCSV(record []string) (t Todo, err error) {
	if l := len(record); l != 4 {
		return Todo{}, fmt.Errorf("invalid length for todo record in CSV format: %v", l)
	}

	t.Description = record[1]

	completed, err := strconv.ParseBool(record[2])
	if err != nil {
		return Todo{}, err
	}
	t.Completed = completed

	unix, err := strconv.ParseInt(record[3], 10, 64)
	if err != nil {
		return Todo{}, err
	}
	t.CreatedAt = time.Unix(unix, 0)

	return
}

func ReadFromCSV(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	if err := lockFile(f); err != nil {
		return err
	}
	defer func() {
		unlockFile(f)
		f.Close()
	}()

	r := csv.NewReader(f)

	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return err
		}

		if _, found := todos[id]; found {
			return fmt.Errorf("duplicate id '%v'", id)
		}

		t, err := NewTodoCSV(record)
		if err != nil {
			return err
		}

		todos[id] = &t
		if id >= nextId {
			nextId = id + 1
		}
	}
}

func WriteToCSV(path string) error {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	if err := lockFile(f); err != nil {
		return err
	}
	defer func() {
		unlockFile(f)
		f.Close()
	}()

	w := csv.NewWriter(f)

	for id, t := range todos {
		r := []string{
			fmt.Sprintf("%v", id),
			t.Description,
			fmt.Sprintf("%v", t.Completed),
			fmt.Sprintf("%v", t.CreatedAt.Unix()),
		}

		if err := w.Write(r); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

func lockFile(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	return nil
}

func unlockFile(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
