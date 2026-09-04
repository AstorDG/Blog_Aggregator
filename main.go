package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/AstorDG/Blog_Aggregator.git/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	current_config, err := read()
	if err != nil {
		println(err)
		return
	}

	db_connection, err := sql.Open("postgres", current_config.DBUrl)
	if err != nil {
		log.Fatal("Coulnd't connect to database")
	}

	db_queries := database.New(db_connection)
	current_state := state{&current_config, db_queries}
	commands := commands{Commands_map: make(map[string]func(*state, command) error, 4)}

	commands.register("login", handler_login)
	commands.register("register", handler_register)
	commands.register("reset", handler_reset)
	commands.register("users", handler_users)
	commands.register("agg", handler_agg)
	commands.register("addfeed", handler_add_feed)

	if len(os.Args) < 2 {
		fmt.Println("No command name given")
		os.Exit(1)
	}

	this_command := command{
		Name:      os.Args[1],
		Arguments: os.Args[2:],
	}
	err = commands.run(&current_state, this_command)
	fmt.Print(err)
}

type command struct {
	Name      string
	Arguments []string
}

type state struct {
	Config   *config
	Database *database.Queries
}

type commands struct {
	Commands_map map[string]func(*state, command) error
}

func (these_commands *commands) run(state_pointer *state, this_command command) error {
	function, key_exists := these_commands.Commands_map[this_command.Name]
	if key_exists {
		return function(state_pointer, this_command)
	}
	return errors.New("Not a valid command")
}

func (these_commands *commands) register(name string, function func(*state, command) error) {
	if _, command_exists := these_commands.Commands_map[name]; command_exists {
		return
	}
	these_commands.Commands_map[name] = function
}
