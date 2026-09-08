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

func handler_add_feed(state_pointer *state, this_command command, this_user database.User) error {
	if len(this_command.Arguments) != 2 {
		log.Fatal("invalid number of arguments. addFeed takes 2 arguments")
	}

	create_feed_args := database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   this_command.Arguments[0],
		Url:    this_command.Arguments[1],
		UserID: this_user.ID,
	}
	feed_database, err := state_pointer.Database.CreateFeed(context.Background(), create_feed_args)
	if err != nil {
		log.Fatal("Couldn't create a feed with that name and url")
	}

	follow_feed_params := database.FollowFeedParams{
		ID:     uuid.New(),
		UserID: this_user.ID,
		FeedID: feed_database.ID,
	}

	state_pointer.Database.FollowFeed(context.Background(), follow_feed_params)

	fmt.Printf("Feed fields: %v", feed_database)
	return nil
}

func handler_feeds(state_pointer *state, this_command command) error {
	if len(this_command.Arguments) > 0 {
		log.Fatal("feeds takes no arguments")
	}

	feeds_from_database, err := state_pointer.Database.GetAllFeeds(context.Background())
	if err != nil {
		log.Fatal("Couldn't get feeds from database")
	}
	for _, feed := range feeds_from_database {
		user_name, err := state_pointer.Database.GetUserNameById(context.Background(), feed.UserID)
		if err != nil {
			user_name = ""
		}
		fmt.Printf("Feed name: %s\n", feed.Name)
		fmt.Printf("Feed url: %s\n", feed.Url)
		fmt.Printf("Feed's User %s\n", user_name)
	}

	return nil
}

func handler_follow(state_pointer *state, this_command command, this_user database.User) error {
	if len(this_command.Arguments) != 1 {
		log.Fatal("Incorrect number of arguments. follow takes one argument")
	}
	url := this_command.Arguments[0]

	feed, err := state_pointer.Database.GetFeedByURL(context.Background(), url)
	if err != nil {
		log.Fatal("Coulnd't a feed with that url")
	}

	follow_feed_params := database.FollowFeedParams{
		ID:     uuid.New(),
		UserID: this_user.ID,
		FeedID: feed.ID,
	}
	follow_feed, err := state_pointer.Database.FollowFeed(context.Background(), follow_feed_params)
	if err != nil {
		log.Fatal("Database error associating user with url")
	}

	fmt.Printf("Feed name: %s\n", follow_feed.FeedName)
	fmt.Printf("User name: %s\n", follow_feed.UserName)
	return nil
}

func handler_following(state_pointer *state, this_command command, this_user database.User) error {
	if len(this_command.Arguments) > 0 {
		log.Fatal("Following doesn't take any arguments")
	}

	user_feeds, err := state_pointer.Database.GetFeedFollowsForUser(context.Background(), this_user.ID)
	if err != nil {
		log.Fatal("Couldn't get feed information about the current user")
	}
	for _, feed := range user_feeds {
		fmt.Printf("Feed name: %s\n", feed.FeedsName)
	}
	return nil
}

func handler_unfollow(state_pointer *state, this_command command, this_user database.User) error {
	if len(this_command.Arguments) != 1 {
		log.Fatal("unfollow only takes a url as an argument")
	}

	feed, err := state_pointer.Database.GetFeedByURL(context.Background(), this_command.Arguments[0])
	if err != nil {
		log.Fatal("Feed with that url doesn't exist in the database")
	}

	unfollow_args := database.UnFollowParams{
		UserID: this_user.ID,
		FeedID: feed.ID,
	}

	err = state_pointer.Database.UnFollow(context.Background(), unfollow_args)
	if err != nil {
		log.Fatal("Error unfollowing that feed. Retry")
	}

	fmt.Printf("%s unfollowed successfully\n", feed.Name)
	return nil
}

func check_logged_in(handler func(state_pointer *state, this_command command, this_user database.User) error) func(*state, command) error {
	return func(state_pointer *state, this_command command) error {
		this_user, err := state_pointer.Database.GetUserByName(context.Background(), state_pointer.Config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(state_pointer, this_command, this_user)
	}
}
