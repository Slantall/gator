package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	cfg, _ := config.Read()
	currentState := &state{
		cfgPointer: &cfg,
	}

	db, err := sql.Open("postgres", currentState.cfgPointer.Dburl)
	dbQueries := database.New(db)
	currentState.db = dbQueries

	cmds := commands{cmdMap: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerFeeds)
	allArgs := os.Args
	if len(allArgs) < 2 {
		fmt.Println("No command given.")
		os.Exit(1)
	}
	cmdName := allArgs[1]
	args := allArgs[2:]
	err = cmds.run(currentState, command{cmdName: cmdName, cmdargs: args})
	if err != nil {
		fmt.Println("Issue running command:", err)
		os.Exit(1)
	}

}
