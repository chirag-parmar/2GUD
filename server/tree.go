package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/m1gwings/treedrawer/tree" // for visualizing merkle trees
	"fmt"
)

type node struct {
	left *node
	right *node
	hash string
}

func (n *node) isLeaf() bool {
	if n.right == nil && n.left == nil {
		return true
	}

	return false
}

func (n *node) areBothChildrenLeaves() bool {
	if n.right.isLeaf() && n.left.isLeaf() {
		return true
	}

	return false
}

type MerkleTree struct {
	root *node
}

func (t *MerkleTree) init() {
	t.root = nil
	t.depth = 0
	t.full = false
	t.numLeaves = 0
	t.searchTable = make(map[string]int)
}

func (t *MerkleTree) AddLeaf(hash string) {
	leaf := node{nil, nil, hash}
	t.numLeaves += 1
	t.searchTable[hash] = t.numLeaves
	
	if t.root == nil {
		t.root = leaf
		t.depth = 0
		t.full = false
		return
	} else if t.root.isLeaf() {
		t.root.left = t.root
		t.root.right = leaf
		t.root.hash = ComputeHash(t.root.left.hash + t.root.right.hash)
		t.depth = 1
		t.full = true
		return
	}

	if t.full {
		t.root.left = t.root
		t.root.right = leaf
		t.root.hash = ComputeHash(t.root.left.hash + t.root.right.hash)
		t.depth += 1
		t.full = false

		return
	}

	curNode := t.root
	var traverseList []*node
	for traverseDepth := range t.depth {
		if curNode.isLeaf() {
			curNode.left = curNode
			curNode.right = leaf
			curNode.hash = ComputeHash(curNode.left.hash + curNode.right.hash)




			if traverseDepth == t.depth - 1 {
				t.full = true
			}
			return
		} else if curNode.areBothChildrenLeaves() {
			curNode.left = curNode
			curNode.right = leaf
			curNode.hash = ComputeHash(curNode.left.hash + curNode.right.hash)
			return
		}
		traverseList = append(traverseList, curNode)
		curNode = curNode.right
	}
}

func (t *MerkleTree) CalculateHash() string {

}
