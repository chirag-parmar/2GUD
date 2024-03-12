package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type node struct {
	left *node
	right *node
	hash string
}

type MerkleTree struct {
	root *node
}

func (t *MerkleTree) AddLeaf(hash string) {
	leaf := new node(nil, nil, hash)
	if t.root.right == nil {
		// tree is empty at root
		t.root = leaf
		return
	}

	curRight := t.root.right
	for curRight != nil {
		curRight = curRight.right
	}
}
