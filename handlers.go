package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AstorDG/Blog_Aggregator.git/internal/database"
	"github.com/google/uuid"
)

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

func handler_agg(state_pointer *state, this_command command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatal("Couln't get a feed from that url")
	}
	fmt.Printf("%v", feed)
	return nil
}

type RSS_item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
type RSS_feed struct {
	Channel struct {
		Title       string     `xml:"title"`
		Link        string     `xml:"link"`
		Description string     `xml:"description"`
		Item        []RSS_item `xml:"item"`
	} `xml:"channel"`
}

func fetchFeed(this_context context.Context, feed_url string) (*RSS_feed, error) {
	new_client := http.DefaultClient
	this_request, err := http.NewRequestWithContext(this_context, "GET", feed_url, nil)
	if err != nil {
		log.Fatal("Couldn't make a request with that shape")
	}
	this_request.Header.Set("User-Agent", "gator")
	response, err := new_client.Do(this_request)
	if err != nil {
		log.Fatal("Error getting response from server")
	}
	defer response.Body.Close()

	response_bytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("couldn't get bytes out of response")
	}

	xml_data := RSS_feed{}
	err = xml.Unmarshal(response_bytes, &xml_data)
	if err != nil {
		return nil, err
	}

	xml_data.Channel.Title = html.UnescapeString(xml_data.Channel.Title)
	xml_data.Channel.Description = html.UnescapeString(xml_data.Channel.Description)
	for index, item := range xml_data.Channel.Item {
		xml_data.Channel.Item[index].Title = html.UnescapeString(item.Title)
		xml_data.Channel.Item[index].Description = html.UnescapeString(item.Description)
	}
	return &xml_data, nil
}
