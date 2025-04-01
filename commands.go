package main

import (
	"context"
	"fmt"
	"gator/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
)

type command struct {
	cmdName string
	cmdargs []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.cmdargs) == 0 {
		return fmt.Errorf("No arguments entered.")
	}
	name := cmd.cmdargs[0]
	if _, err := s.db.GetUser(context.Background(), name); err != nil {
		fmt.Printf("User '%s' does not exist\n", name)
		os.Exit(1)
	}
	s.cfgPointer.SetUser(name)
	fmt.Printf("User %v has been sent to config\n", name)
	return nil
}

type commands struct {
	cmdMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmdMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	foundCmd, exists := c.cmdMap[cmd.cmdName]
	if !exists {
		return fmt.Errorf("command %q not found", cmd.cmdName)
	}
	err := foundCmd(s, cmd)
	return err
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.cmdargs) == 0 {
		return fmt.Errorf("No name for user.")
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.cmdargs[0],
	}

	if _, err := s.db.GetUser(context.Background(), params.Name); err == nil {
		fmt.Printf("User '%s' already exists\n", params.Name)
		os.Exit(1)
	}
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Failed to create user: %w", err)
	}
	s.cfgPointer.SetUser(params.Name)
	fmt.Printf("User %v created.\n", params.Name)
	fmt.Printf("User details: %v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to create user: %w", err)
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to retrieve user: %w", err)
	}
	for _, user := range users {
		if s.cfgPointer.CurrentUserName == user {
			user = fmt.Sprintf("%v (current)", user)
		}
		fmt.Println(user)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.cmdargs) != 2 {
		return fmt.Errorf("Incorrect amount of args entered. 'AddFeed' requires 2 args: 'name' and 'url'")
	}

	url := cmd.cmdargs[1]
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.cmdargs[0],
		Url:       url,
		UserID:    currentUser.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Println(feed)
	followcmd := command{
		cmdName: "follow",
		cmdargs: []string{url},
	}
	handlerFollow(s, followcmd, currentUser)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to retrieve feeds: %w", err)
	}
	for _, feed := range feeds {
		userName, err := s.db.GetUserWithID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("Failed to retrieve user: %w", err)
		}
		fmt.Println("Name:", feed.Name, "URL:", feed.Url, "Feed Creator:", userName.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.cmdargs) < 1 {
		return fmt.Errorf("URL argument is required")
	}
	url := cmd.cmdargs[0]
	feed, err := s.db.GetFeedWithURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Failed to retrieve feeds: %w", err)
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}
	followResult, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Failed to create feed follow: %w", err)
	}
	fmt.Printf("%s is now following %s\n", followResult.UserName, followResult.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, currentUser database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("Failed to find following list %w", err)
	}
	if len(following) == 0 {
		fmt.Println("You aren't following any feeds.")
		return nil
	}
	for i, feed := range following {
		fmt.Printf("%d. %s\n", i+1, feed.FeedName)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfgPointer.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Failed to retrieve user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.cmdargs) < 1 {
		return fmt.Errorf("URL argument is required")
	}
	params := database.UnfollowParams{
		UserID: currentUser.ID,
		Url:    cmd.cmdargs[0],
	}
	err := s.db.Unfollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Failed to unfollow: %w", err)
	}
	return nil
}
