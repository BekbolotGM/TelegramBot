package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestGetTokenAndDevKey(t *testing.T) {
	bottoken, devkey := GetTokenAndDevKey()
	assert.NotEqual(t, "", bottoken)
	assert.NotEqual(t, "", devkey)
}

func TestSearchingVideo(t *testing.T) {
	_, devkey := GetTokenAndDevKey()
	text, err := SearchingVideo("hello", devkey)
	require.NoError(t, err)
	assert.NotEqual(t, "", text)
	fmt.Println(text)
}

func TestConvertingVideo(t *testing.T) {
	_, devkey := GetTokenAndDevKey()
	text, _ := SearchingVideo("hello", devkey)
	ok, err := ConvertingVideo(text)
	require.NoError(t, err)
	require.Equal(t, true, ok)
}
