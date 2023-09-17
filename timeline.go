package main

import (
	"context"
	"encoding/json"
	"fmt"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"strings"
)

func doShowPost(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()
	if !strings.HasPrefix(arg, "https://bsky.app/profile/") && !strings.Contains(arg, "/post/") {
		return fmt.Errorf("uri does not start with 'https://bsky.app/profile/ and /post/': %s", arg)
	}
	arg = strings.TrimPrefix(arg, "https://bsky.app/profile/")
	ids := strings.Split(arg, "/post/")
	if len(ids) != 2 {
		return fmt.Errorf("uri does not contain 2 parts: %s", arg)
	}
	handle := ids[0]
	postID := ids[1]

	ctx := context.Background()
	idResolve, err := comatproto.IdentityResolveHandle(ctx, xrpcc, handle)
	//phr := &api.ProdHandleResolver{}
	//out, err := phr.ResolveHandleToDid(ctx, handle)
	if err != nil {
		return err
	}
	out := idResolve.Did
	did := "at://" + out + "/app.bsky.feed.post/" + postID
	fmt.Println(did)
	dids := []string{did}
	resp, err := bsky.FeedGetPosts(context.TODO(), xrpcc, dids)
	if err != nil {
		return fmt.Errorf("cannot get post thread: %w", err)
	}
	if len(resp.Posts) == 0 {
		return fmt.Errorf("no posts found")
	}
	printPost(resp.Posts[0])
	return nil
}

func doThread(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()
	if !strings.HasPrefix(arg, "at://did:plc:") {
		arg = "at://did:plc:" + arg
	}

	n := cCtx.Int64("n")
	resp, err := bsky.FeedGetPostThread(context.TODO(), xrpcc, n, n, arg)
	if err != nil {
		return fmt.Errorf("cannot get post thread: %w", err)
	}

	replies := resp.Thread.FeedDefs_ThreadViewPost.Replies
	if cCtx.Bool("json") {
		err = json.NewEncoder(os.Stdout).Encode(resp.Thread.FeedDefs_ThreadViewPost)
		if err != nil {
			return err
		}
		for _, p := range replies {
			json.NewEncoder(os.Stdout).Encode(p)
		}
		return nil
	}

	for i := 0; i < len(replies)/2; i++ {
		replies[i], replies[len(replies)-i-1] = replies[len(replies)-i-1], replies[i]
	}
	printPost(resp.Thread.FeedDefs_ThreadViewPost.Post)
	for _, r := range replies {
		printPost(r.FeedDefs_ThreadViewPost.Post)
	}
	return nil
}

func execTimeline(cCtx *cli.Context, handle string, limit int64) (feed []*bsky.FeedDefs_FeedViewPost, err error) {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}

	var cursor string

	for {
		if handle != "" {
			if handle == "self" {
				handle = xrpcc.Auth.Did
			}
			resp, err := bsky.FeedGetAuthorFeed(context.TODO(), xrpcc, handle, cursor, "", limit)
			if err != nil {
				return nil, fmt.Errorf("cannot get author feed: %w", err)
			}
			feed = append(feed, resp.Feed...)
			if resp.Cursor != nil {
				cursor = *resp.Cursor
			} else {
				cursor = ""
			}
		} else {
			resp, err := bsky.FeedGetTimeline(context.TODO(), xrpcc, "reverse-chronological", cursor, limit)
			if err != nil {
				return nil, fmt.Errorf("cannot get timeline: %w", err)
			}
			feed = append(feed, resp.Feed...)
			if resp.Cursor != nil {
				cursor = *resp.Cursor
			} else {
				cursor = ""
			}
		}
		if cursor == "" || int64(len(feed)) > limit {
			break
		}
	}

	sort.Slice(feed, func(i, j int) bool {
		ri := timep(feed[i].Post.Record.Val.(*bsky.FeedPost).CreatedAt)
		rj := timep(feed[j].Post.Record.Val.(*bsky.FeedPost).CreatedAt)
		return ri.Before(rj)
	})
	if int64(len(feed)) > limit {
		feed = feed[len(feed)-int(limit):]
	}
	return feed, nil
}

func doTimeline(cCtx *cli.Context) (err error) {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	n := cCtx.Int64("n")
	handle := cCtx.String("handle")

	feed, err := execTimeline(cCtx, handle, n)

	if cCtx.Bool("json") {
		for _, p := range feed {
			json.NewEncoder(os.Stdout).Encode(p)
		}
	} else {
		for _, p := range feed {
			//if p.Reason != nil {
			//continue
			//}
			printPost(p.Post)
		}
	}

	return nil
}
