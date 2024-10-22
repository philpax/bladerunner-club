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
	botType := GetBotResponseType(post.Text)

	if botType == -1 {
		return nil
	}

	authorDID := c.Account.Identity.DID

	if post.Reply == nil || IsSelfThread(c, post) {
		mentionedDids := mentionedDids(post)
		for _, botDID := range mentionedDids {
			handleBotSignal(c, botDID, authorDID, botType)
		}
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
	handleBotSignal(c, botDID, authorDID, botType)
	return nil
}

func handleBotSignal(c *automod.RecordContext, botDID string, authorDID string, botType int) {
	if botType == 1 {
		c.IncrementDistinct("goodbot", botDID, authorDID)
		c.IncrementDistinct("bladerunner", authorDID, botDID)
		c.Logger.Error("good bot reply")

		// XXX: bypass counts for early testing
		if err = addAccountLabel(c, botDID, "good-bot"); err != nil {
			return err
		}

		if c.GetCountDistinct("goodbot", botDID, countstore.PeriodTotal) > GOOD_BOT_REPLY_THRESHOLD-1 {
			c.Logger.Error("good bot")
			// c.AddAccountLabel("good-bot")
			// c.Notify("slack")
		}

		if c.GetCountDistinct("bladerunner", authorDID, countstore.PeriodTotal) > BLADERUNNER_THRESHOLD-1 {
			c.Logger.Error("bladerunner")
			// c.AddAccountLabel("bladerunner")
			// c.Notify("slack")
		}

		return
	}

	c.IncrementDistinct("badbot", botDID, authorDID)
	c.IncrementDistinct("jabroni", authorDID, botDID)
	c.Logger.Error("bad bot reply")

	if c.GetCountDistinct("badbot", botDID, countstore.PeriodTotal) > BAD_BOT_REPLY_THRESHOLD-1 {
		// @TODO: this would add label to the reply author's account not the parent/bot's account
		// c.AddAccountLabel("bad-bot")
		c.Logger.Error("bad bot")
		// c.Notify("slack")
	}

	if c.GetCountDistinct("jabroni", authorDID, countstore.PeriodTotal) > JABRONI_THRESHOLD-1 {
		// c.AddAccountLabel("jabroni")
		c.Logger.Error("jabroni")
		// c.Notify("slack")
	}
}

func mentionedDids(post *appbsky.FeedPost) []string {
	mentionedDids := []string{}
	for _, facet := range post.Facets {
		for _, feature := range facet.Features {
			mention := feature.RichtextFacet_Mention
			if mention == nil {
				continue
			}
			mentionedDids = append(mentionedDids, mention.Did)
		}
	}

	return mentionedDids
}

// @TODO: this is a dumb check that only matches text exactly, could be improved
func GetBotResponseType(s string) int {
	// Normalize the string by converting to lowercase and trimming spaces
	tokens := TokenizeText(strings.TrimSpace(strings.ToLower(s)))
	hasGoodBot := false
	hasBadBot := false

	for i, token := range tokens {
		if token != "good" && token != "bad" && token != "bot" && !strings.HasPrefix(token, "@") {
			return -1
		}

		if (token == "good" || token == "bad") && i+1 < len(tokens) && tokens[i+1] == "bot" {
			if token == "good" {
				hasGoodBot = true
			} else {
				hasBadBot = true
			}
		}
	}

	if hasGoodBot {
		return 1
	}

	if hasBadBot {
		return 0
	}

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

func TokenizeText(text string) []string {
	return strings.Fields(text)
}
