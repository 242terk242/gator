package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
	"github.com/mik242/gator/internal/config"
	"github.com/mik242/gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
	// add more commands here
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *commands) run(s *state, cmd command) error {
	if handler, ok := c.handlers[cmd.name]; ok {
		return handler(s, cmd)
	}
	return errors.New("unknown command: " + cmd.name)
}
func (c *commands) register(name string, f func(*state, command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
}

func handlerLogin(state *state, cmd command) error {
	if len(cmd.args) < 1 {
		fmt.Println("need at least one argument for login")
		os.Exit(1)
	}
	_, err := state.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println("User not found, please register first.")
		os.Exit(1)
	}
	err = state.cfg.SetUser(cmd.args[0])
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func handlerRegister(state *state, cmd command) error {
	if len(cmd.args) < 1 {
		fmt.Println("need at least one argument for register")
		os.Exit(1)
	}
	state.cfg.CurrentUserName = cmd.args[0]
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	_, err := state.db.CreateUser(context.Background(), params)
	if err != nil {
		return err.(error)
	}
	err = state.cfg.SetUser(cmd.args[0])
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func handlerReset(state *state, cmd command) error {
	err := state.db.DeleteUsers(context.Background())
	if err != nil {
		fmt.Println("Error deleting user:", err)
		os.Exit(1)
	}
	return nil
}

func handlerUsers(state *state, cmd command) error {
	users, err := state.db.ListUsers(context.Background())
	if err != nil {
		return err.(error)
	}
	for _, user := range users {
		if state.cfg.CurrentUserName == user.Name {
			fmt.Println(user.Name + " (current)")
		} else {
			fmt.Println(user.Name)
		}
	}
	return nil
}

func handlerAgg(state *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("need at least one argument for agg")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Println("Starting aggregation with interval:", cmd.args[0])
	fmt.Printf("Collecting feeds every %s\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		scrapeFeeds(state)
	}

}

func handlerScrape(cmd command) error {

	feedURL := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return err.(error)
	}
	RSSFeed := feed
	fmt.Printf("Feed Title: %s\n", RSSFeed.Channel.Title)
	fmt.Printf("Feed Link: %s\n", RSSFeed.Channel.Link)
	fmt.Printf("Feed Description: %s\n", RSSFeed.Channel.Description)
	for _, item := range RSSFeed.Channel.Item {
		fmt.Printf("Item Title: %s\n", item.Title)
		fmt.Printf("Item Link: %s\n", item.Link)
		fmt.Printf("Item Description: %s\n", item.Description)
		fmt.Printf("Item PubDate: %s\n", item.PubDate)
	}
	return nil
}

func handlerAddFeed(state *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		fmt.Println("need at least two arguments for addFeed")
		os.Exit(1)
	}
	if state.cfg.CurrentUserName == "" {
		return errors.New("no user is currently logged in")
	}

	user, err := state.db.GetUser(context.Background(), state.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found: %w", err)
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID,
	}
	feed, err := state.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return err
	}
	fmt.Println("Feed added successfully.", feedParams.ID, feedParams.CreatedAt, "Name:", feedName, "URL:", feedURL, feedParams.UserID)

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = state.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	return nil
}

func handlerFeeds(state *state, cmd command) error {
	if state.cfg.CurrentUserName == "" {
		return errors.New("no user is currently logged in")
	}

	result, err := state.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("current user not found: %w", err)
	}

	for _, feed := range result {
		fmt.Printf("Feed Name: %s\n", feed.Name)
		fmt.Printf("Feed URL: %s\n", feed.Url)
		fmt.Printf("Feed Created At: %s\n", feed.CreatedAt)
		fmt.Printf("Feed Updated At: %s\n", feed.UpdatedAt)
		fmt.Printf("Feed User Name: %s\n", feed.UserName)
	}
	return nil
}

func handlerFollow(state *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("need at least one argument for follow")
		os.Exit(1)
	}
	if state.cfg.CurrentUserName == "" {
		return errors.New("no user is currently logged in")
	}

	user, err := state.db.GetUser(context.Background(), state.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found: %w", err)
	}

	feed, err := state.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	feedID := feed.ID
	if feedID == uuid.Nil {
		return fmt.Errorf("feed not found: %s", cmd.args[0])
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedID,
	}
	_, err = state.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}
	fmt.Println("Successfully followed the feed.")
	return nil
}

func handlerFollowing(state *state, cmd command, user database.User) error {
	if state.cfg.CurrentUserName == "" {
		return errors.New("no user is currently logged in")
	}

	user, err := state.db.GetUser(context.Background(), state.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found: %w", err)
	}

	follows, err := state.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows: %w", err)
	}

	for _, follow := range follows {
		fmt.Printf("Feed Follow Name: %s\n", follow.FeedName)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("feed request failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	return &feed, nil
}

func handlerUnfollow(state *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("need at least one argument for unfollow")
		os.Exit(1)
	}
	if state.cfg.CurrentUserName == "" {
		return errors.New("no user is currently logged in")
	}

	feed, err := state.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	feedID := feed.ID
	if feedID == uuid.Nil {
		return fmt.Errorf("feed not found: %s", cmd.args[0])
	}

	return state.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	})
}

func scrapeFeeds(state *state) error {
	//Get the next feed to fetch from the DB.
	// Mark it as fetched.
	// Fetch the feed using the URL (we already wrote this function)
	// Iterate over the items in the feed and print their titles to the console.

	nextFeed, err := state.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %w", err)
	}

	err = state.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %w", err)
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	for _, item := range feed.Channel.Item {
		saveParams := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			FeedID:    nextFeed.ID,
			Title:     item.Title,
			URL:       item.Link,
			Content:   item.Description,
		}
		_, err := state.db.CreatePost(context.Background(), saveParams)
		if err != nil {
			return fmt.Errorf("failed to save post: %w", err)
		}
		fmt.Printf("Saved Post Title: %s\n", item.Title)
	}
	return nil
}

func handlerBrowse(state *state, cmd command) error {
	posts, err := state.db.GetPosts(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("Post Title: %s\n", post.Title)
		fmt.Printf("Post URL: %s\n", post.URL)
		fmt.Printf("Post Content: %s\n", post.Content)
		fmt.Printf("Post Feed Name: %s\n", post.FeedName)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return errors.New("no user is currently logged in")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("current user not found: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func main() {
	var args = os.Args
	if len(args) < 2 {
		fmt.Println("Please provide a command.")
		os.Exit(1)
	}
	var cmds commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", handlerBrowse)
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	dbURL := cfg.DbUrl
	if dbURL == "" {
		fmt.Println("DATABASE_URL environment variable must be set")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	var s state
	s = state{cfg: cfg, db: database.New(db)}

	if err := cmds.run(&s, command{name: args[1], args: args[2:]}); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}
