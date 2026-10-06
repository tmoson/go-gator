package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/tmoson/go-gator/internal/config"
	"github.com/tmoson/go-gator/internal/database"
)

func main() {
	conf := config.Read()
	db, err := sql.Open("postgres", conf.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	state := State{
		db:   dbQueries,
		conf: &conf,
	}
	commands := Commands{
		commands: make(map[string]func(*State, Command) error),
	}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerGetFeeds)
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Too few arguments passed, expected 2, but got %d\n", len(args))
		os.Exit(1)
	}
	command := Command{
		name: args[1],
		args: args[2:],
	}
	err = commands.run(&state, command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
