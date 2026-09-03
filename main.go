package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AstorDG/Blog_Aggregator.git/internal/database"
	"github.com/google/uuid"
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

func handler_login(state_pointer *state, this_command command) error {
	if len(this_command.Arguments) == 0 {
		fmt.Println("a username is required")
		os.Exit(1)
	}
	if len(this_command.Arguments) != 1 {
		return errors.New("too many arguments")
	}
	this_user_name := this_command.Arguments[0]
	_, err := state_pointer.Database.GetUserByName(context.Background(), this_user_name)
	if err != nil {
		log.Fatal("User doesn't exist")
	}
	state_pointer.Config.set_user(this_user_name)
	fmt.Printf("user has been set to: %v", this_user_name)
	return nil
}

func handler_register(state_pointer *state, this_command command) error {
	if len(this_command.Arguments) != 1 {
		log.Fatal("Incorrect number of arguments passed to register")
	}
	new_user_name := this_command.Arguments[0]
	new_db_user, err := state_pointer.Database.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: new_user_name})
	if err != nil {
		log.Fatal("User with that name already exists")
	}
	state_pointer.Config.set_user(new_user_name)
	fmt.Printf("user created with name: %s", new_user_name)
	log.Println(new_db_user)
	return nil
}

func handler_reset(state_pointer *state, this_command command) error {
	err := state_pointer.Database.ResetDatabase(context.Background())
	if err != nil {
		log.Fatal("Error reseting database")
	}
	return nil
}

func handler_users(state_pointer *state, this_command command) error {
	all_user_names, err := state_pointer.Database.GetUsers(context.Background())
	if err != nil {
		log.Fatal("Couldn't get users")
	}
	for _, user := range all_user_names {
		if user == state_pointer.Config.CurrentUserName {
			fmt.Printf("%v (current)\n", user)
		} else {
			fmt.Println(user)
		}
	}
	return nil
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
