package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// TestTreeBuilder builds a tree and gets the merkle root.
func TestTreeBuilder(t *testing.T) {
	tree := NewTree()
	tree.PushObject("a")
	tree.PushObject("b")
	_ = tree.Root()

	// Correctness is assumed, as it's tested by the merkletree package. This
	// function is really for code coverage.
}

// TestCalculateLeaves probes the CalulateLeaves function.
func TestCalculateLeaves(t *testing.T) {
	tests := []struct {
		size, expSegs uint64
	}{
		{0, 1},
		{63, 1},
		{64, 1},
		{65, 2},
		{127, 2},
		{128, 2},
		{129, 3},
	}

	for i, test := range tests {
		if segs := CalculateLeaves(test.size); segs != test.expSegs {
			t.Errorf("miscalculation for test %v: expected %v, got %v", i, test.expSegs, segs)
		}
	}
}

// TestStorageProof builds a storage proof and checks that it verifies
// correctly.
func TestStorageProof(t *testing.T) {
	// Generate proof data.
	numSegments := uint64(7)
	data := make([]byte, numSegments*SegmentSize)
	rand.Read(data)
	rootHash, err := ReaderMerkleRoot(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	// Create and verify proofs for all indices.
	for i := uint64(0); i < numSegments; i++ {
		baseSegment, hashSet, err := BuildReaderProof(bytes.NewReader(data), i)
		if err != nil {
			t.Error(err)
			continue
		}
		if !VerifySegment(baseSegment, hashSet, numSegments, i, rootHash) {
			t.Error("Proof", i, "did not pass verification")
		}
	}

	// Try an incorrect proof.
	baseSegment, hashSet, err := BuildReaderProof(bytes.NewReader(data), 3)
	if err != nil {
		t.Fatal(err)
	}
	if VerifySegment(baseSegment, hashSet, numSegments, 4, rootHash) {
		t.Error("Verified a bad proof")
	}
}

// TestNonMultipleNumberOfSegmentsStorageProof builds a storage proof that has
// a last leaf of size less than SegmentSize.
func TestNonMultipleLeafSizeStorageProof(t *testing.T) {
	// Generate proof data.
	data := make([]byte, (2*SegmentSize)+10)
	rand.Read(data)
	rootHash, err := ReaderMerkleRoot(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	// Create and verify a proof for the last index.
	baseSegment, hashSet, err := BuildReaderProof(bytes.NewReader(data), 2)
	if err != nil {
		t.Error(err)
	}
	if !VerifySegment(baseSegment, hashSet, 3, 2, rootHash) {
		t.Error("padded segment proof failed")
	}
}

// TestCachedTree tests the cached tree functions of the package.
func TestCachedTree(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Build a cached tree out of 4 subtrees, each subtree of height 2 (4
	// elements).
	tree1Bytes, err := RandBytes(SegmentSize * 4)
	if err != nil {
		t.Fatal(err)
	}
	tree2Bytes, err := RandBytes(SegmentSize * 4)
	if err != nil {
		t.Fatal(err)
	}
	tree3Bytes, err := RandBytes(SegmentSize * 4)
	if err != nil {
		t.Fatal(err)
	}
	tree4Bytes, err := RandBytes(SegmentSize * 4)
	if err != nil {
		t.Fatal(err)
	}
	tree1Root, err := ReaderMerkleRoot(bytes.NewReader(tree1Bytes))
	if err != nil {
		t.Fatal(err)
	}
	tree2Root, err := ReaderMerkleRoot(bytes.NewReader(tree2Bytes))
	if err != nil {
		t.Fatal(err)
	}
	tree3Root, err := ReaderMerkleRoot(bytes.NewReader(tree3Bytes))
	if err != nil {
		t.Fatal(err)
	}
	tree4Root, err := ReaderMerkleRoot(bytes.NewReader(tree4Bytes))
	if err != nil {
		t.Fatal(err)
	}
	fullRoot, err := ReaderMerkleRoot(bytes.NewReader(append(tree1Bytes, append(tree2Bytes, append(tree3Bytes, tree4Bytes...)...)...)))
	if err != nil {
		t.Fatal(err)
	}

	// Get a cached proof for index 0.
	base, cachedHashSet, err := BuildReaderProof(bytes.NewReader(tree1Bytes), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifySegment(base, cachedHashSet, 4, 0, tree1Root) {
		t.Fatal("the proof for the subtree was invalid")
	}
	ct := NewCachedTree(2)
	ct.SetIndex(0)
	ct.Push(tree1Root)
	ct.Push(tree2Root)
	ct.Push(tree3Root)
	ct.Push(tree4Root)
	hashSet := ct.Prove(base, cachedHashSet)
	if !VerifySegment(base, hashSet, 4*4, 0, fullRoot) {
		t.Fatal("cached proof construction appears unsuccessful")
	}
	if ct.Root() != fullRoot {
		t.Fatal("cached Merkle root is not matching the full Merkle root")
	}

	// Get a cached proof for index 6.
	base, cachedHashSet, err = BuildReaderProof(bytes.NewReader(tree2Bytes), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifySegment(base, cachedHashSet, 4, 2, tree2Root) {
		t.Fatal("the proof for the subtree was invalid")
	}
	ct = NewCachedTree(2)
	ct.SetIndex(6)
	ct.Push(tree1Root)
	ct.Push(tree2Root)
	ct.Push(tree3Root)
	ct.Push(tree4Root)
	hashSet = ct.Prove(base, cachedHashSet)
	if !VerifySegment(base, hashSet, 4*4, 6, fullRoot) {
		t.Fatal("cached proof construction appears unsuccessful")
	}
	if ct.Root() != fullRoot {
		t.Fatal("cached Merkle root is not matching the full Merkle root")
	}
}
