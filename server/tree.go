package main

import (
	"math"
)

type node struct {
	left *node
	right *node
	weight int
	hash string
}

func (n *node) isLeaf() bool {
	if n.right == nil && n.left == nil {
		return true
	}

	return false
}

func (n *node) hasEqualChildren() bool {
	if n.right.weight == n.left.weight {
		return true
	}

	return false
}

func (n *node) addLeaf(hash string) {

	if n.isLeaf() || n.hasEqualChildren() {
		// move the existing tree to the left and include a right
		n = &node{
			left: n, 
			right: &node{nil, nil, 1, hash}
		}

		n.weight = n.left.weight + n.right.weight
		n.hash = ComputeHash(n.left.hash + n.right.hash)

		return
	}

	n.right.AddLeaf(hash)
	n.weight = n.left.weight + n.right.weight
	n.hash = ComputeHash(n.left.hash + n.right.hash)

	return
}

type MerkleTree struct {
	root *node
}

func (t *MerkleTree) AddLeaf(hash string) {
	t.root.addLeaf(hash)
	return
}

func (t *MerkleTree) GetDepth() int {
	return math.Ceil(math.Log2(t.root.weight))
}
