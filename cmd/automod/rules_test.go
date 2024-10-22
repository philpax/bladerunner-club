package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBotResponseType(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(GetBotResponseType("bad bot"), 0)
	assert.Equal(GetBotResponseType("good bot"), 1)
	assert.Equal(GetBotResponseType("good bot beahvior is punished"), -1)
}
