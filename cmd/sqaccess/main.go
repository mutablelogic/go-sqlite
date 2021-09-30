package main

import (
	"fmt"

	// Namespace Imports
	//. "github.com/mutablelogic/go-sqlite"

	// Modules
	"github.com/mutablelogic/go-sqlite/pkg/sqobj"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type User struct {
	Name string `sqlite:"name,primary"`
	Hash string `sqlite:"hash,not null"`
}

func main() {
	fmt.Println("IN")
	if r, err := sqobj.NewReflect(User{}); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}
