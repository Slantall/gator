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
