package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tmoson/go-gator/internal/config"
	"github.com/tmoson/go-gator/internal/database"
)

type State struct {
	db   *database.Queries
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
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return errors.New(fmt.Sprintf("User not recognized, please register %s as a user first", cmd.args[0]))
	}
	err = s.conf.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User set to: %s", cmd.args[0])
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if cmd.hasNoArgs() {
		return errors.New("Need a username to register.")
	}
	now := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.args[0],
	})
	if err != nil {
		return err
	}
	s.conf.SetUser(cmd.args[0])
	fmt.Printf(
		"%s was created, id: %v, created_at: %v, updated_at: %v, name: %s\n",
		cmd.args[0],
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Name,
	)
	return nil
}

func handlerUsers(s *State, cmd Command) error {
	currentUser := s.conf.CurrentUserName
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerReset(s *State, cmd Command) error {
	err := s.db.ResetUsers(context.Background())
	return err
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
