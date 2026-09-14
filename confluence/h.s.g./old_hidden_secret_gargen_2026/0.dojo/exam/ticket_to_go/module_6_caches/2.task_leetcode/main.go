package main

import "fmt"

/*
LeetCode 208 — Trie (prefix tree).
Insert / Search / StartsWith — O(len(word)).
*/
type TrieNode struct {
	children [26]*TrieNode
	end      bool
}

type Trie struct{ root *TrieNode }

func Constructor() Trie { return Trie{root: &TrieNode{}} }

func (t *Trie) Insert(word string) {
	n := t.root
	for i := 0; i < len(word); i++ {
		idx := word[i] - 'a'
		if n.children[idx] == nil {
			n.children[idx] = &TrieNode{}
		}
		n = n.children[idx]
	}
	n.end = true
}

func (t *Trie) Search(word string) bool {
	n := t.walk(word)
	return n != nil && n.end
}

func (t *Trie) StartsWith(prefix string) bool {
	return t.walk(prefix) != nil
}

func (t *Trie) walk(s string) *TrieNode {
	n := t.root
	for i := 0; i < len(s); i++ {
		idx := s[i] - 'a'
		if n.children[idx] == nil {
			return nil
		}
		n = n.children[idx]
	}
	return n
}

func main() {
	tr := Constructor()
	tr.Insert("apple")
	fmt.Println(tr.Search("apple"))   // true
	fmt.Println(tr.Search("app"))     // false
	fmt.Println(tr.StartsWith("app")) // true
	tr.Insert("app")
	fmt.Println(tr.Search("app")) // true
}
