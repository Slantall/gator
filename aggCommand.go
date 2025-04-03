package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, err
	}

	// Unescape HTML entities in the channel fields
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	// Unescape HTML entities in each item's fields
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.cmdargs) < 1 {
		return fmt.Errorf("duration string is required")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.cmdargs[0])
	if err != nil {
		return fmt.Errorf("Failed to parse time between requests duration: %w", err)
	}
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Fprintf(os.Stderr, "Error scraping feeds: %v\n", err)
		}
	}

	//return nil
}

func scrapeFeeds(s *state) error {
	nextfeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to find Feed to fetch: %w", err)
	}

	currentTime := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	markFetchedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: currentTime,
		ID:            nextfeed.ID,
	}

	err = s.db.MarkFeedFetched(context.Background(), markFetchedParams)
	if err != nil {
		return fmt.Errorf("Failed to mark feed as fetched: %w", err)
	}
	feed, err := fetchFeed(context.Background(), nextfeed.Url)
	if err != nil {
		return fmt.Errorf("Failed to fetch RSS Feed: %w", err)
	}

	fmt.Printf("Feed: %s\n", feed.Channel.Title)

	for _, i := range feed.Channel.Item {
		desc := sql.NullString{
			String: i.Description,
			Valid:  true,
		}
		pubTime, err := time.Parse(time.RFC1123, i.PubDate)
		if err != nil {
			return fmt.Errorf("Failed to parse publish time: %w", err)
		}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       i.Title,
			Url:         i.Link,
			Description: desc,
			PublishedAt: pubTime,
			FeedID:      nextfeed.ID,
		}
		post, err := s.db.CreatePost(context.Background(), postParams)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					continue
				} else {
					return fmt.Errorf("Failed to create post, due to PostgreSQL error: %w", err)
				}
			} else {
				return fmt.Errorf("Failed to create post, due to error: %w", err)
			}
		}
		fmt.Printf("Creating post: %s\n", post.Title)
	}

	return nil
}
