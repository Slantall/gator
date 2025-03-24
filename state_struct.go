package main

import (
	"gator/internal/config"
	"gator/internal/database"
)

type state struct {
	db         *database.Queries
	cfgPointer *config.Config
}
