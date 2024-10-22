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

	botDID, err := parentURI.Authority().AsDID()
	if err != nil {
		return err
	}
	authorDID := c.Account.Identity.DID

	if botType == 1 {
		c.IncrementDistinct("goodbot", botDID.String(), authorDID.String())
		c.IncrementDistinct("bladerunner", authorDID.String(), botDID.String())
		c.Logger.Error("good bot reply")

		// XXX: bypass counts for early testing
		if err = addAccountLabel(c, botDID, "good-bot"); err != nil {
			return err
		}

		if c.GetCountDistinct("goodbot", botDID.String(), countstore.PeriodTotal) > GOOD_BOT_REPLY_THRESHOLD-1 {
			c.Logger.Error("good bot")
			// c.AddAccountLabel("good-bot")
			// c.Notify("slack")
		}

		if c.GetCountDistinct("bladerunner", authorDID.String(), countstore.PeriodTotal) > BLADERUNNER_THRESHOLD-1 {
			c.Logger.Error("bladerunner")
			// c.AddAccountLabel("bladerunner")
			// c.Notify("slack")
		}

		return nil
	}

	c.IncrementDistinct("badbot", botDID.String(), authorDID.String())
	c.IncrementDistinct("jabroni", authorDID.String(), botDID.String())
	c.Logger.Error("bad bot reply")

	if c.GetCountDistinct("badbot", botDID.String(), countstore.PeriodTotal) > BAD_BOT_REPLY_THRESHOLD-1 {
		// @TODO: this would add label to the reply author's account not the parent/bot's account
		// c.AddAccountLabel("bad-bot")
		c.Logger.Error("bad bot")
		// c.Notify("slack")
	}

	if c.GetCountDistinct("jabroni", authorDID.String(), countstore.PeriodTotal) > JABRONI_THRESHOLD-1 {
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
