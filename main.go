package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

type Todo struct {
	Description string
	Completed   bool
	CreatedAt   time.Time
}

var todos map[int]Todo = make(map[int]Todo)
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
	default:
		fmt.Println("invalid argument")
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
	nextId++
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

func (t Todo) Add() {
	todos[nextId] = t
	nextId++
}

func ReadFromCSV(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

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

		todos[id] = t
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
	defer f.Close()

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
