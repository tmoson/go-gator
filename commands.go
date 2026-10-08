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

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
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

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("Need a name and url to register feed.")
	}
	now := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf(
		"%s was created, id: %v, created_at: %v, updated_at: %v, url: %s\n",
		feed.Name,
		feed.ID,
		feed.CreatedAt,
		feed.UpdatedAt,
		feed.Url,
	)
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func handlerGetFeeds(s *State, cmd Command) error {
	if !cmd.hasNoArgs() {
		return errors.New("feeds does not take arguments")
	}
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for i := range feeds {
		feed := feeds[i]
		fmt.Printf(
			"Feed: %s\nURL: %s\nCreated By %s At: %v, Updated: %v\n",
			feed.Name,
			feed.Url,
			feed.UserName.String,
			feed.CreatedAt,
			feed.UpdatedAt,
		)
	}
	return nil
}

func handlerFollow(s *State, cmd Command, user database.User) error {
	if cmd.hasNoArgs() {
		return errors.New("Need url to follow feed")
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	now := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s followed %s", user.Name, feed.Name)
	return nil
}

func handlerFollowing(s *State, cmd Command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for i := range feeds {
		fmt.Printf("%s\n", feeds[i])
	}
	return nil
}

func handlerUnfollow(s *State, cmd Command, user database.User) error {
	if cmd.hasNoArgs() {
		return errors.New("Need feed to unfollow, received none")
	}
	return s.db.UserUnfollowFeed(context.Background(), database.UserUnfollowFeedParams{
		UserID: user.ID,
		Url:    cmd.args[0],
	})
}

func handlerAgg(s *State, cmd Command) error {
	rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	// fmt.Printf("Feed: %s\n", rssFeed.Channel.Link)
	// fmt.Printf("Title: %s\n", rssFeed.Channel.Title)
	// fmt.Printf("Description: %s\n", rssFeed.Channel.Description)
	// for i := 0; i < len(rssFeed.Channel.Item); i++ {
	// 	rssItem := rssFeed.Channel.Item[i]
	// 	fmt.Printf("\nTitle: %s\n", rssItem.Title)
	// 	fmt.Printf("%s\nDescription: %s\n", rssItem.PubDate, rssItem.Description)
	// 	fmt.Printf("Link: %s\n", rssItem.Link)
	// }
	fmt.Printf("%v", rssFeed)
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
