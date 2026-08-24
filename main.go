package main

import (
	"fmt"
	"os"

	"github.com/tmoson/go-gator/internal/config"
)

func main() {
	conf := config.Read()
	state := State{
		conf: &conf,
	}
	commands := Commands{
		commands: make(map[string]func(*State, Command) error),
	}
	commands.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Too few arguments passed, expected 2, but got %d\n", len(args))
		os.Exit(1)
	}
	command := Command{
		name: args[1],
		args: args[2:],
	}
	err := commands.run(&state, command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
