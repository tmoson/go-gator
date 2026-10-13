package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

func parseDate(s string) (time.Time, error) {
	dateLayouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00",
		"Mon, 12 Oct 2026 10:27:00 +0000",
	}
	for _, layout := range dateLayouts {
		// copying this from example, want to simplify nil err checks
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New(fmt.Sprintf("Date %s not in any standard format", s))
}

func scrapeFeeds(s *State) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("Feed: %s\n", rssFeed.Channel.Link)
	fmt.Printf("Title: %s\n", rssFeed.Channel.Title)
	fmt.Printf("Description: %s\n", rssFeed.Channel.Description)
	for i := 0; i < len(rssFeed.Channel.Item); i++ {
		rssItem := rssFeed.Channel.Item[i]
		pubDate, err := parseDate(rssItem.PubDate)
		if err != nil {
			return err
		}
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			Title: sql.NullString{
				String: rssItem.Title,
				Valid:  true,
			},
			Url: rssItem.Link,
			Description: sql.NullString{
				String: rssItem.Description,
				Valid:  true,
			},
			PublishedAt: pubDate,
			FeedID:      nextFeed.ID,
		})
	}
	return nil
}

func handlerAgg(s *State, cmd Command) error {
	if cmd.hasNoArgs() {
		return errors.New("Need duration to run requests in, e.g. 1s, 1m, 1h")
	}
	interval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	fmt.Printf("Fetching feeds every %s\n", cmd.args[0])
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	var err error
	if !cmd.hasNoArgs() {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("%s\nPublished: %v\n\n%s\n\n", post.Title.String, post.PublishedAt, post.Description.String)
	}
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
