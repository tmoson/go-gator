package main

import (
	"errors"
	"fmt"

	"github.com/tmoson/go-gator/internal/config"
)

type State struct {
	conf *config.Config
}

type Command struct {
	name string
	args []string
}

func (c Command) hasNoArgs() bool {
	return len(c.args) == 0
}

func handlerLogin(s *State, cmd Command) error {
	if cmd.hasNoArgs() {
		return errors.New("Not enough arguments to login. Login expects 1 argument, but received 0.")
	}
	err := s.conf.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User set to: %s", cmd.args[0])
	return nil
}

type Commands struct {
	commands map[string]func(*State, Command) error
}

func (c *Commands) run(s *State, cmd Command) error {
	command, exists := c.commands[cmd.name]
	if !exists {
		return fmt.Errorf("Invalid command %s", cmd.name)
	}
	err := command(s, cmd)
	return err
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.commands[name] = f
}
