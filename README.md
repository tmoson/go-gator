# go-gator

This project is a simple rss feed aggregator written in go. It makes use of a postgresql database, meant to be run locally, to store users, feeds, follows, and posts.

## Installation

There are a few prerequisite tools that you need to install in order to run go-gator. Install these first:
1. [PostgreSQL](https://www.postgresql.org/docs/current/tutorial-install.html) (The database the program uses)
2. [Goose](https://github.com/pressly/goose) (The migration tool used to update the project)
3. [Golang](https://go.dev/doc/install) (I'm currently on 1.27.2 when writing this)

Once you have the tools installed, [initialize your database](https://www.postgresql.org/docs/current/app-initdb.html), and use goose to create the tables.

With an initialized db, create a centralized config file at `~/.gatorconfig.json`, with `"db_url"` set to the url of your application. This application currently has no security set up for protecting keys, so the username and password will likely be stored in plaintext, as opposed to something secure. So you may have a `.gatorconfig.json that looks like this:

```json
{
    "db_url": "postgres://username:password@localhost:1234/gator?sslmode=disable"
}
```

Once your db is initialized and your config set, go ahead and run `go build` followed by `go install`. If you haven't insalled any go tools before this, you may want to make sure that the go install path is a part of your `PATH` by following the [compile and install docs](https://go.dev/doc/tutorial/compile-install) from the golang website.

## Running the program

Once installed, you just need to use `go-gator` to run one of the following commands:

|  Command  |                  Arguments                   |                                Output                               |
| --------- | -------------------------------------------- | ------------------------------------------------------------------- |
| register  | username                                     | confirmation that user was created                                  |
| login     | username                                     | notification that you are logged in as user                         |
| users     | none                                         | list of all users with current user indicated                       |
| addfeed   | feedName, feedUrl                            | confirmation that feed was created                                  |
| feeds     | none                                         | a list of all feeds                                                 |
| agg       | refreshInterval (e.g. 5s, 1m, 1h)            | the refresh interval and the name of feeds before requests are sent |
| follow    | feedUrl                                      | confirmation that current user has followed url                     |
| following | none                                         | a list of all the feeds the current user is following               |
| unfollow  | feedUrl                                      | confirmation that feed has been unfollowed                          |
| browse    | number of posts (optional)                   | the last n posts from current user's feeds, defaulting to 2         |
| reset     | none                                         | none                                                                |

Feeds cannot be created without users, and posts cannot be seen without feeds, so you will need to create user(s), have users create feeds, and then run the `agg` command on its own in a terminal where it won't interfere with users. The `agg` command runs continually until it is stopped, and if you're not careful can DDOS your feeds.
