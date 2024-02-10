package nexns

import (
	"encoding/json"
	"testing"
)

func TestTrieInsert(t *testing.T) {
	trie := &Trie{root: &TrieNode{}}

	domainJsonData := `{
		"domain": {
			"id": 1, "domain": "example.com",
			"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
			"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
		},
		"zones": [{
			"id": 11, "name": "default", "rules": ["0.0.0.0/0"],
			"rrsets": [
				{ "id": 111, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.0.1"}]},
				{ "id": 112, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.2"}]},
				{ "id": 113, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.3"}]}
			]
		}]
	}`
	domainData := &DomainData{}
	json.Unmarshal([]byte(domainJsonData), domainData)
	trie.Insert(domainData)

	domainJsonData2 := `{
		"domain": {
			"id": 2, "domain": "sub.example.com",
			"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
			"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
		},
		"zones": [{
			"id": 11, "name": "default", "rules": ["0.0.0.0/0"],
			"rrsets": [
				{ "id": 111, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.0.1"}]},
				{ "id": 112, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.2"}]},
				{ "id": 113, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.3"}]}
			]
		}]
	}`
	domainData2 := &DomainData{}
	json.Unmarshal([]byte(domainJsonData2), domainData2)
	trie.Insert(domainData2)

	domainJsonData3 := `{
		"domain": {
			"id": 3, "domain": "test.com",
			"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
			"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
		},
		"zones": [{
			"id": 11, "name": "default", "rules": ["0.0.0.0/0"],
			"rrsets": [
				{ "id": 111, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.0.1"}]},
				{ "id": 112, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.2"}]},
				{ "id": 113, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.3"}]}
			]
		}]
	}`
	domainData3 := &DomainData{}
	json.Unmarshal([]byte(domainJsonData3), domainData3)
	trie.Insert(domainData3)

	domainJsonData4 := `{
		"domain": {
			"id": 4, "domain": "top",
			"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
			"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
		},
		"zones": [{
			"id": 11, "name": "default", "rules": ["0.0.0.0/0"],
			"rrsets": [
				{ "id": 111, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.0.1"}]},
				{ "id": 112, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.2"}]},
				{ "id": 113, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.3"}]}
			]
		}]
	}`
	domainData4 := &DomainData{}
	json.Unmarshal([]byte(domainJsonData4), domainData4)
	trie.Insert(domainData4)
}

func buildTestingTrie() (*Trie, error) {
	/*
		Build a testing trie (only use A record)

		Contents:

		com
			example: www 1.0.0.1, ftp 1.0.0.2, sub.www 1.0.0.3
				sub: www 1.0.1.1, ftp 1.0.1.2, sub.www 1.0.1.3
			test: www 1.1.0.1, ftp 1.1.0.2, sub.www 1.1.0.3
		top: www 2.0.0.1, ftp 2.0.0.2, sub.www 2.0.0.3

	*/
	domainJsonData := `[
		{
			"domain": {
				"id": 1, "domain": "example.com",
				"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
				"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
			},
			"zones": [{
				"id": 11, "name": "default", "rules": ["0.0.0.0/0"],
				"rrsets": [
					{ "id": 111, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.0.1"}]},
					{ "id": 112, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.2"}]},
					{ "id": 113, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.0.3"}]}
				]
			}]
		},
		{
			"domain": {
				"id": 2, "domain": "sub.example.com",
				"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
				"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
			},
			"zones": [{
				"id": 21, "name": "default", "rules": ["0.0.0.0/0"],
				"rrsets": [
					{ "id": 211, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.0.1.1"}]},
					{ "id": 212, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.1.2"}]},
					{ "id": 213, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.0.1.3"}]}
				]
			}]
		},
		{
			"domain": {
				"id": 3, "domain": "test.com",
				"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
				"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
			},
			"zones": [{
				"id": 31, "name": "default", "rules": ["0.0.0.0/0"],
				"rrsets": [
					{ "id": 311, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "1.1.0.1"}]},
					{ "id": 312, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.1.0.2"}]},
					{ "id": 313, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "1.1.0.3"}]}
				]
			}]
		},
		{
			"domain": {
				"id": 4, "domain": "top",
				"mname": "ns.example.com", "rname": "root.example.com", "serial": "123456789",
				"refresh": 3600, "retry": 3600, "expire": 3600, "ttl": 3600
			},
			"zones": [{
				"id": 41, "name": "default", "rules": ["0.0.0.0/0"],
				"rrsets": [
					{ "id": 411, "name": "www", "type": "A", "records": [{"id": 1, "ttl": 3600, "val": "2.0.0.1"}]},
					{ "id": 412, "name": "ftp", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "2.0.0.2"}]},
					{ "id": 413, "name": "sub.www", "type": "A", "records": [{"id": 2, "ttl": 3600, "val": "2.0.0.3"}]}
				]
			}]
		}
	]`
	domainJsonDataBytes := []byte(domainJsonData)
	var domainData []DomainData
	// domainData := make([]DomainData, 0)
	err := json.Unmarshal(domainJsonDataBytes, &domainData)

	trie := BuildTrie(domainData)
	return trie, err
}

// test if trie built successfully
func TestTrieBuild(t *testing.T) {
	trie, err := buildTestingTrie()
	if err != nil {
		t.Fatalf("Error building test trie: %s", err)
	}

	if _, exists := trie.root.children["com"]; !exists {

		keys := make([]string, 0, len(trie.root.children))
		for k := range trie.root.children {
			keys = append(keys, k)
		}

		t.Logf("root.children map keys: %s", keys)
		t.Fatalf("Failed to build trie node `com`")
	}

	if _, exists := trie.root.children["com"].children["example"]; !exists {

		keys := make([]string, 0, len(trie.root.children))
		for k := range trie.root.children {
			keys = append(keys, k)
		}

		t.Logf("root.children[com].children map keys: %s", keys)
		t.Fatalf("Failed to build trie node `example.com`")
	}

	if trie.root.children["com"].children["example"].domainData.Domain.ID != 1 {
		t.Fatalf("Failed to build domain `example.com`")
	}
	if trie.root.children["com"].children["example"].children["sub"].domainData.Domain.ID != 2 {
		t.Fatalf("Failed to build domain `sub.example.com`")
	}
	if trie.root.children["com"].children["test"].domainData.Domain.ID != 3 {
		t.Fatalf("Failed to build domain `test.com`")
	}
	if trie.root.children["top"].domainData.Domain.ID != 4 {
		t.Fatalf("Failed to build domain `top`")
	}
}

func TestTrieSearch(t *testing.T) {
	trie, err := buildTestingTrie()
	if err != nil {
		t.Fatalf("Error building test trie: %s", err)
	}

	domainData := trie.Search("example.com")
	if domainData.Domain.ID != 1 {
		t.Fatalf("Failed to search domain `example.com`")
	}

	domainData = trie.Search("sub.example.com")
	if domainData.Domain.ID != 2 {
		t.Fatalf("Failed to search domain `sub.example.com`")
	}

	domainData = trie.Search("test.com")
	if domainData.Domain.ID != 3 {
		t.Fatalf("Failed to search domain `test.com`")
	}

	domainData = trie.Search("top")
	if domainData.Domain.ID != 4 {
		t.Fatalf("Failed to search domain `top`")
	}
}
