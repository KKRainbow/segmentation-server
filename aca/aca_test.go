package aca_test

import (
	"testing"
	"unicode/utf8"

	. "github.com/naturali/kmr/pkg/aca"

	"github.com/magiconair/properties/assert"
)

func TestEmpty(t *testing.T) {
	ac := NewAhoCorasickMatcher()
	ac.Build(nil)
}

func TestInsert(t *testing.T) {
	ac := NewAhoCorasickMatcher()
	ac.Build([]string{"say", "she", "shell", "shr", "her"})
	rs := []rune("aaashellaashrmmmmmhemmhera")
	ret, pos := ac.MatchRunes(rs)
	for i := 0; i < len(ret); i++ {
		assert.Equal(t, ret[i], string(rs[pos[i]:pos[i]+utf8.RuneCountInString(ret[i])]))
	}
}

func TestChineseInsert(t *testing.T) {
	ac := NewAhoCorasickMatcher()
	ac.Build([]string{"hi", "北京", "北京天安门", "人民广场"})
	rs := []rune("从北京天安门去人民广场去say hi")
	ret, pos := ac.MatchRunes(rs)
	for i := 0; i < len(ret); i++ {
		assert.Equal(t, ret[i], string(rs[pos[i]:pos[i]+utf8.RuneCountInString(ret[i])]))
	}
	assert.Equal(t, len(ret), 4)
}

func TestCoverInsert(t *testing.T) {
	ac := NewAhoCorasickMatcher()
	ac.Build([]string{"北京", "广场", "北京广场", "广场北京"})
	rs := []rune("北京广场北京广场")
	ret, pos := ac.MatchRunes(rs)
	for i := 0; i < len(ret); i++ {
		assert.Equal(t, ret[i], string(rs[pos[i]:pos[i]+utf8.RuneCountInString(ret[i])]))
	}
	assert.Equal(t, len(ret), 7)
}

func TestDuplicateInsert(t *testing.T) {
	ac := NewAhoCorasickMatcher()
	ac.Build([]string{"aa"})
	rs := []rune("aaaaa")
	ac.MatchRunes(rs)
	ret, pos := ac.MatchRunes(rs)
	for i := 0; i < len(ret); i++ {
		assert.Equal(t, ret[i], string(rs[pos[i]:pos[i]+utf8.RuneCountInString(ret[i])]))
	}
	assert.Equal(t, len(ret), 4)
}
