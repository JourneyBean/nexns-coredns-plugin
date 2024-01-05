package mewwoof

import (
	"strings"
)

// TrieNode represents a node in the Trie
type TrieNode struct {
	domainData *DomainData
	children   map[string]*TrieNode
}

// Trie represents the Trie data structure
type Trie struct {
	root *TrieNode
}

// Insert inserts a domain into the Trie
func (t *Trie) Insert(domainData *DomainData) {
	node := t.root
	labels := strings.Split(domainData.Domain.Name, ".")

	// reverse and walk domain
	for i := range labels {
		label := labels[len(labels)-1-i]

		if node.children == nil {
			node.children = make(map[string]*TrieNode)
		}

		if _, exists := node.children[label]; !exists {
			node.children[label] = &TrieNode{}
		}

		node = node.children[label]
	}

	node.domainData = domainData
}

// Search searches for a domain in the Trie and returns the corresponding DomainData
// 从根开始，最长匹配
func (t *Trie) Search(domain string) *DomainData {

	// remove "." suffix for FQDN
	if domain[len(domain)-1] == '.' {
		domain = domain[:len(domain)-1]
	}

	node := t.root
	labels := strings.Split(domain, ".")

	// reverse and walk domain
	for i := range labels {
		label := labels[len(labels)-1-i]

		// 1' is leaf, no longer matches
		// db: foo.example.com, query: bar.foo.example.com => foo.example.com
		if node.children == nil {
			// has data
			if node.domainData != nil {
				return node.domainData
			}
			return nil
		}

		// 2' no longer matches
		// db: example.com, foo.example.com, query: bar.example.com => example.com
		if _, exists := node.children[label]; !exists {
			// has data
			if node.domainData != nil {
				return node.domainData
			}
			return nil
		}

		node = node.children[label]
	}

	// 3' exactly match
	// db: foo.example.com, query: foo.example.com => foo.example.com
	if node.domainData != nil {
		return node.domainData
	}

	return nil
}

// delete, domain must be exact
func (t *Trie) Delete(domain string) {
	// remove "." suffix for FQDN
	if domain[len(domain)-1] == '.' {
		domain = domain[:len(domain)-1]
	}

	node := t.root
	labels := strings.Split(domain, ".")
	visitPath := make([]*TrieNode, 0)

	// reverse and walk domain
	for i := range labels {
		label := labels[len(labels)-1-i]

		visitPath = append(visitPath, node)

		// 1' db: example.com, domain: foo.example.com => invalid
		if node.children == nil {
			return
		}

		// 2' db: foo.example.com, domain: bar.example.com => invalid
		if _, exists := node.children[label]; !exists {
			return
		}

		node = node.children[label]
	}
	visitPath = append(visitPath, node)

	// 3' exact match
	if node.domainData != nil {
		node.domainData = nil
	}

	// chain delete map
	for i := range labels {
		childrenNode := visitPath[len(visitPath)-1-i]
		parentNode := visitPath[len(visitPath)-2-i]
		parentLabel := labels[i]

		if (childrenNode.children == nil || len(childrenNode.children) == 0) && childrenNode.domainData == nil {
			delete(parentNode.children, parentLabel)
		}
	}
}

// BuildTrie builds a Trie from a list of DomainData
func BuildTrie(data []DomainData) *Trie {
	trie := &Trie{root: &TrieNode{}}

	for _, entry := range data {
		trie.Insert(&entry)
	}

	return trie
}
