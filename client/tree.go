package main

import (
	"math"
	"strings"
	"fmt"
	"errors"
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

func (n *node) bothChildrenAreLeaves() bool {
	if n.right.isLeaf() && n.left.isLeaf() {
		return true
	}

	return false
}

func (n *node) addLeaf(hash string) {

	if n.isLeaf() || n.hasEqualChildren() {
		// move the existing tree to the left and include a right
		n.left = &node{n.left, n.right, n.weight, n.hash} 
		n.right = &node{nil, nil, 1, hash}

		// new node update
		n.weight = n.left.weight + n.right.weight
		n.hash = ComputeHash(n.left.hash + n.right.hash)

		return
	}

	n.right.addLeaf(hash)

	// parent node update
	n.weight = n.left.weight + n.right.weight
	n.hash = ComputeHash(n.left.hash + n.right.hash)

	return
}

type MerkleTree struct {
	root *node
	numLeaves int
	indexToHash map[int]string
	hashToIndex map[string]int
}

func (t *MerkleTree) Init(hash string) {
	t.root = &node{nil, nil, 1, hash}
	t.numLeaves = 1
	t.indexToHash = make(map[int]string)
	t.indexToHash[0] = hash
	t.hashToIndex = make(map[string]int)
	t.hashToIndex[hash] = 0
}

func (t *MerkleTree) AddLeaf(hash string) error {
	t.indexToHash[t.numLeaves] = hash
	t.hashToIndex[hash] = t.numLeaves
	t.numLeaves += 1
	t.root.addLeaf(hash)

	// count serves as a additional measure for sanctity of tree update
	if t.numLeaves != t.root.weight {
		return errors.New("Something went wrong with the tree update!")
	}

	return nil
}

func (t *MerkleTree) Depth() int {
	return int(math.Ceil(math.Log2(float64(t.root.weight))))
}

func (t *MerkleTree) GetProofByIndex(index int) []string {
	path := fmt.Sprintf("%b", index)
	
	// FIXME: There is probably a way to avoid this
	if len(path) < t.Depth() {
		path = strings.Repeat("0", t.Depth() - len(path)) + path
	}

	var proof []string
	curNode := t.root
	for i := 0; i < len(path); i++ {
		if curNode.isLeaf() {
			break
		}

		if (i < len(path) - 1) && curNode.bothChildrenAreLeaves() {
			continue
		}

		if path[i] == '0' {
			proof = append(proof, "0") //instruction for concatenation during verification
			proof = append(proof, curNode.right.hash)
			curNode = curNode.left
		} else {
			proof = append(proof, "1") //instruction for concatenation during verification
			proof = append(proof, curNode.left.hash)
			curNode = curNode.right
		}
	}

	return proof
}

func (t *MerkleTree) GetProofByHash(hash string) []string {
	path := fmt.Sprintf("%b", t.hashToIndex[hash])
	
	// FIXME: There is probably a way to avoid this
	if len(path) < t.Depth() {
		path = strings.Repeat("0", t.Depth() - len(path)) + path
	}

	var proof []string
	curNode := t.root
	for i := 0; i < len(path); i++ {
		if curNode.isLeaf() {
			break
		}

		if (i < len(path) - 1) && curNode.bothChildrenAreLeaves() {
			continue
		}

		if path[i] == '0' {
			proof = append(proof, "0") //instruction for concatenation during verification
			proof = append(proof, curNode.right.hash)
			curNode = curNode.left
		} else {
			proof = append(proof, "1") //instruction for concatenation during verification
			proof = append(proof, curNode.left.hash)
			curNode = curNode.right
		}
	}

	return proof
}

func VerifyProof(content string, proof []string, rootHash string) bool {
	computedHash := ComputeHash(content)

	for i := len(proof)-1; i >= 0; i -= 2 {
		if proof[i-1] == "0" {
			computedHash = ComputeHash(computedHash + proof[i])
		} else {
			computedHash = ComputeHash(proof[i] + computedHash)
		}
	}

	if rootHash != computedHash {
		return false
	}

	return true
}


// // Quick and Dirty testing
// func main() {
// 	t := MerkleTree{}
// 	t.Init(ComputeHash("A"))
// 	t.AddLeaf(ComputeHash("B"))

// 	// checkValue1 := ComputeHash(ComputeHash("A") + ComputeHash("B"))

// 	// if t.root.hash != checkValue1 {
// 	// 	fmt.Println("Fail1")
// 	// 	return
// 	// }

// 	t.AddLeaf(ComputeHash("C"))

// 	// checkValue2 := ComputeHash(checkValue1 + ComputeHash("C"))

// 	// if t.root.hash != checkValue2 {
// 	// 	fmt.Println("Fail2")
// 	// 	return
// 	// }

// 	t.AddLeaf(ComputeHash("D"))

// 	// checkValue3 := ComputeHash(checkValue1 + ComputeHash(ComputeHash("C") + ComputeHash("D")))

// 	// if t.root.hash != checkValue3 {
// 	// 	fmt.Println("Fail2")
// 	// 	return
// 	// }

// 	t.AddLeaf(ComputeHash("E"))
// 	t.AddLeaf(ComputeHash("F"))
// 	t.AddLeaf(ComputeHash("G")) 

// 	// temp := ComputeHash(ComputeHash(ComputeHash("E") + ComputeHash("F")) + ComputeHash("G"))
// 	// checkValue4 := ComputeHash(checkValue3 + temp)

// 	// if t.root.hash != checkValue4 {
// 	// 	fmt.Println("Fail3")
// 	// 	return
// 	// }

// 	t.AddLeaf(ComputeHash("H")) 

// 	// temp = ComputeHash(ComputeHash(ComputeHash("E") + ComputeHash("F")) + ComputeHash(ComputeHash("G") + ComputeHash("H")))
// 	// checkValue5 := ComputeHash(checkValue3 + temp)

// 	// if t.root.hash != checkValue5 {
// 	// 	fmt.Println("Fail5")
// 	// }

// 	if !VerifyProof("C", t.GetProofByHash(ComputeHash("C")), t.root.hash) {
// 		fmt.Println("Proof fail!")
// 		return
// 	}

// 	if !VerifyProof("C", t.GetProofByIndex(2), t.root.hash) {
// 		fmt.Println("Proof fail!")
// 		return
// 	}

// 	fmt.Println("Pass!!!")
// 	return
// }