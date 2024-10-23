package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePostText(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(parsePostText("bad bot"), NilBot)
	assert.Equal(parsePostText("good bot"), GoodBot)
	assert.Equal(parsePostText("good bot beahvior is punished"), BadBot)
	assert.Equal(parsePostText("testing good bot one"), BadBot)
	assert.Equal(parsePostText("testing good bot"), BadBot)
	assert.Equal(parsePostText("@test.bsky.social good bot"), GoodBot)
	assert.Equal(parsePostText("bad bot @one.bsky.social"), NilBot)
}
