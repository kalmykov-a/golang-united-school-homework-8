package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var (
	errorOperationFlag = errors.New("-operation flag has to be specified")
	errorFilenameFlag  = errors.New("-fileName flag has to be specified")
	errorItemFlag      = errors.New("-item flag has to be specified")
	errorIdFlag        = errors.New("-id flag has to be specified")
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	if fileName != "" {
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			return fmt.Errorf("cannot open file:%w", err)
		}
		defer file.Close()
	} else {
		return errorFilenameFlag
	}
	operation, okOperation := args["operation"]
	if okOperation {
		switch operation {
		case "add":
			item := args["item"]
			if item == "" {
				return errorItemFlag
			} else {
				return add(item, fileName, writer)
			}
		case "list":
			return list(fileName, writer)
		case "findById":
			id := args["id"]
			if id != "" {
				return findById(id, fileName, writer)
			} else {
				return errorIdFlag
			}
		case "remove":
			id := args["id"]
			if id != "" {
				return remove(id, fileName, writer)
			} else {
				return errorIdFlag
			}
		case "":
			return errorOperationFlag
		default:
			return fmt.Errorf("Operation %s not allowed!", operation)
		}
	} else {
		return errorOperationFlag
	}
	return nil
}

func parseArgs() Arguments {
	operationFlag := flag.String("operation", "", "")
	itemFlag := flag.String("item", "", "")
	fileNameFlag := flag.String("fileName", "", "")
	idFlag := flag.String("id", "", "")
	flag.Parse()

	return Arguments{
		"operation": *operationFlag,
		"item":      *itemFlag,
		"fileName":  *fileNameFlag,
		"id":        *idFlag,
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func list(fileName string, writer io.Writer) error {
	bytes, err := os.ReadFile(fileName) // For read access.
	if err != nil {
		return fmt.Errorf("cannot read file bytes:%w", err)
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return fmt.Errorf("cannot write data:%w", err)
	}
	return nil
}

func add(item string, fileName string, writer io.Writer) error {
	users, _ := getStruct(fileName)
	u := User{}
	err := json.Unmarshal([]byte(item), &u)
	if err != nil {
		return fmt.Errorf("cannot unmarshal file:%w", err)
	}
	for _, v := range users {
		if v.Id == u.Id {
			_, err := writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", u.Id)))
			if err != nil {
				return fmt.Errorf("cannon write data:%w", err)
			}
		}

	}
	users = append(users, u)
	err = writeStruct(fileName, users)
	if err != nil {
		return fmt.Errorf("cannot write stuct:%w", err)
	}
	return nil
}

func findById(id string, fileName string, writer io.Writer) error {
	users, err := getStruct(fileName)
	if err != nil {
		return fmt.Errorf("cannot get struct from file: %w", err)
	}
	for i, v := range users {
		if id == v.Id {
			bytes, err := json.Marshal(users[i])
			if err != nil {
				return fmt.Errorf("cannot marshal: %w", err)
			}
			_, err = writer.Write(bytes)
			if err != nil {
				return fmt.Errorf("cannot write: %w", err)
			}
		}
	}
	return nil
}

func remove(id string, fileName string, writer io.Writer) error {
	users, err := getStruct(fileName)
	if err != nil {
		return fmt.Errorf("Cannot get struct from file: %w", err)
	}
	var newUsers []User
	for _, v := range users {
		if id != v.Id {
			newUsers = append(newUsers, v)
			err := writeStruct(fileName, newUsers)
			if err != nil {
				return fmt.Errorf("cannot write stuct:%w", err)
			}
			break
		}
	}
	if len(users) == len(newUsers) {
		_, err := writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		if err != nil {
			return fmt.Errorf("cannon write data:%w", err)
		}
	}
	return nil
}

func getStruct(filename string) ([]User, error) {
	file, err := os.Open(filename)
	var u []User
	if err != nil {
		return nil, fmt.Errorf("cannot open file:%w", err)
	}
	bytes, err := ioutil.ReadAll(file)
	err = json.Unmarshal(bytes, &u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse data from JSON:%w", err)
	}
	return u, nil
}

func writeStruct(filename string, u []User) error {
	jsonStr, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("error marshalling:%w", err)
	}
	err = ioutil.WriteFile(filename, jsonStr, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write json:%w", err)
	}
	return nil
}
