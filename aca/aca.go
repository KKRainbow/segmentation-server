package aca

import (
	"unicode/utf8"
)

type trieNode struct {
	nexts map[rune]*trieNode
	value string
	f     *trieNode
}

func newTrieNode() *trieNode {
	return &trieNode{
		nexts: make(map[rune]*trieNode),
	}
}

func (t *trieNode) move(r rune) *trieNode {
	if v, ok := t.nexts[r]; ok {
		return v
	} else {
		return nil
	}
}

func (t *trieNode) addChild(r rune) {
	if _, ok := t.nexts[r]; !ok {
		t.nexts[r] = newTrieNode()
	}
}

func (t *trieNode) setValue(value string) {
	t.value = value
}

func (t *trieNode) setFail(node *trieNode) {
	t.f = node
}

func (t *trieNode) walkBuildTrie(seq string) *trieNode {
	tr := t
	for _, r := range seq {
		tr.addChild(r)
		tr = tr.move(r)
	}
	return tr
}

type AhoCorasickMatcher struct {
	root *trieNode
}

func NewAhoCorasickMatcher() *AhoCorasickMatcher {
	return &AhoCorasickMatcher{}
}

func (m *AhoCorasickMatcher) Build(keywords []string) {
	m.root = newTrieNode()
	for _, s := range keywords {
		node := m.root.walkBuildTrie(s)
		node.setValue(s)
		node.value = s
	}
	queue := make([]*trieNode, 0)
	queue = append(queue, m.root)
	for len(queue) > 0 {
		cur := queue[0]
		for ch, child := range cur.nexts {
			if cur == m.root {
				child.setFail(m.root)
			} else {
				suffixMatchNode := cur.f
				for suffixMatchNode != nil {
					if _, ok := suffixMatchNode.nexts[ch]; ok {
						break
					}
					suffixMatchNode = suffixMatchNode.f
				}
				if suffixMatchNode == nil {
					child.setFail(m.root)
				} else {
					child.setFail(suffixMatchNode.nexts[ch])
				}
			}
			queue = append(queue, child)
		}
		queue = queue[1:]
	}
}

func (m *AhoCorasickMatcher) MatchRunes(rs []rune) ([]string, []int) {
	ret := make([]string, 0)
	pos := make([]int, 0)
	node := m.root
	for i, r := range rs {
		if v, ok := node.nexts[r]; ok {
			node = v
		} else {
			suffixMatchNode := node.f
			for suffixMatchNode != nil {
				if _, ok := suffixMatchNode.nexts[r]; ok {
					break
				}
				suffixMatchNode = suffixMatchNode.f
			}
			if suffixMatchNode == nil {
				node = m.root
			} else {
				node = suffixMatchNode.nexts[r]
			}
		}
		for t := node; t != nil; t = t.f {
			if t.value != "" {
				ret = append(ret, t.value)
				pos = append(pos, i-utf8.RuneCountInString(t.value)+1)
			}
		}
	}
	return ret, pos
}

func (m *AhoCorasickMatcher) Match(b string) ([]string, []int) {
	return m.MatchRunes([]rune(b))
}
