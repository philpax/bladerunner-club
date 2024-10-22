package main

import (
	"strings"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/automod"
	"github.com/bluesky-social/indigo/automod/countstore"
)

var _ automod.PostRuleFunc = GoodbotBadbotRule
var GOOD_BOT_REPLY_THRESHOLD = 2
var BAD_BOT_REPLY_THRESHOLD = 2
var BLADERUNNER_THRESHOLD = 2
var JABRONI_THRESHOLD = 2

func GoodbotBadbotRule(c *automod.RecordContext, post *appbsky.FeedPost) error {
	if post.Reply == nil || IsSelfThread(c, post) {
		return nil
	}

	botType := GetBotResponseType(post.Text)

	if botType == -1 {
		return nil
	}

	parentURI, err := syntax.ParseATURI(post.Reply.Parent.Uri)
	if err != nil {
		return nil
	}

	botDid := parentURI.Authority().String()
	authorDid := c.Account.Identity.DID.String()

	if botType == 1 {
		c.IncrementDistinct("goodbot", botDid, authorDid)
		c.IncrementDistinct("bladerunner", authorDid, botDid)
		c.Logger.Error("good bot reply")

		if c.GetCountDistinct("goodbot", botDid, countstore.PeriodTotal) > GOOD_BOT_REPLY_THRESHOLD-1 {
			c.Logger.Error("good bot")
			// c.AddAccountLabel("good-bot")
			// c.Notify("slack")
		}

		if c.GetCountDistinct("bladerunner", authorDid, countstore.PeriodTotal) > BLADERUNNER_THRESHOLD-1 {
			c.Logger.Error("bladerunner")
			// c.AddAccountLabel("bladerunner")
			// c.Notify("slack")
		}

		return nil
	}

	c.IncrementDistinct("badbot", botDid, authorDid)
	c.IncrementDistinct("jabroni", authorDid, botDid)
	c.Logger.Error("bad bot reply")

	if c.GetCountDistinct("badbot", botDid, countstore.PeriodTotal) > BAD_BOT_REPLY_THRESHOLD-1 {
		// @TODO: this would add label to the reply author's account not the parent/bot's account
		// c.AddAccountLabel("bad-bot")
		c.Logger.Error("bad bot")
		// c.Notify("slack")
	}

	if c.GetCountDistinct("jabroni", authorDid, countstore.PeriodTotal) > JABRONI_THRESHOLD-1 {
		// c.AddAccountLabel("jabroni")
		c.Logger.Error("jabroni")
		// c.Notify("slack")
	}

	return nil
}

// @TODO: this isn't a clean check and doing some duplicate regex checks that can be avoided
func GetBotResponseType(s string) int {
	// Normalize the string by converting to lowercase and trimming spaces
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "good bot" {
		return 1
	}

	if s == "bad bot" {
		return 0
	}

	// If neither pattern matches
	return -1
}

// checks if the post event is a reply post for which the author is replying to themselves, or author is the root author (OP)
func IsSelfThread(c *automod.RecordContext, post *appbsky.FeedPost) bool {
	if post.Reply == nil {
		return false
	}
	did := c.Account.Identity.DID.String()
	parentURI, err := syntax.ParseATURI(post.Reply.Parent.Uri)
	if err != nil {
		return false
	}
	rootURI, err := syntax.ParseATURI(post.Reply.Root.Uri)
	if err != nil {
		return false
	}

	if parentURI.Authority().String() == did || rootURI.Authority().String() == did {
		return true
	}
	return false
}
