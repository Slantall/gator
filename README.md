# gator
Postgres and Go are both required to be installed to run the program.
Install the gator CLI using go install.
Manually create a config file in your home directory, ~/.gatorconfig.json, with the following content:
{
  "db_url": "postgres://example"
}
Your connection string will be in this format: protocol://username:password@host:port/database

# Commands
    register: register a user. Takes one value after the command for the user's name.
	login: Login to a registered user. Takes one value after the command for the user's name.
    reset: Resets all users.
    agg: Aggregates the followed feeds. Takes one time value (1s, 1m, 1h) for refresh period.
    addfeed: creates and follows a feed. Takes two values: feed 'name' and feed 'url'.
    feeds: Prints all the feeds in the database. 
    follow: Follows a feed. Takes one value after the command: feed 'url'.
    following: List all feeds the current user is following.
    unfollow: Unfollows a feed. Takes one value after the command: feed 'url'.
    browse: Shows the posts of the feeds the user follows. Takes one optional value after the command: limit. (default 2)
