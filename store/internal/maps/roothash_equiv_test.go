package maps

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
)

// TestRootHashFromMapMatchesProofsFromMap is a consensus-safety guard for the
// CommitInfo.Hash optimization: CommitInfo.Hash switched from ProofsFromMap
// (which builds + discards full merkle proofs) to RootHashFromMap. The app
// hash MUST stay byte-identical, so assert the two agree (and that the
// merkleMap-based HashFromMap agrees too) across many shapes of input.
func TestRootHashFromMapMatchesProofsFromMap(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	mkRandom := func(n, keyLen, valLen int) map[string][]byte {
		m := make(map[string][]byte, n)
		for len(m) < n {
			k := make([]byte, keyLen)
			v := make([]byte, valLen)
			r.Read(k)
			r.Read(v)
			m[string(k)] = v
		}
		return m
	}

	// Realistic CommitInfo.toMap() shape: store names -> 32-byte commit hashes.
	storeInfoLike := func(n int) map[string][]byte {
		m := make(map[string][]byte, n)
		for i := 0; i < n; i++ {
			h := sha256.Sum256([]byte(fmt.Sprintf("store-%d", i)))
			m[fmt.Sprintf("module%02d", i)] = h[:]
		}
		return m
	}

	var cases []map[string][]byte
	cases = append(cases, map[string][]byte{"a": {0x01}})                        // single entry
	cases = append(cases, storeInfoLike(1), storeInfoLike(2), storeInfoLike(40)) // realistic
	for _, n := range []int{1, 2, 3, 5, 8, 17, 33, 50} {
		cases = append(cases, mkRandom(n, 1+r.Intn(20), 1+r.Intn(64)))
	}

	for i, m := range cases {
		root := RootHashFromMap(m)
		proofRoot, _, _ := ProofsFromMap(m)
		if !bytes.Equal(root, proofRoot) {
			t.Fatalf("case %d (n=%d): RootHashFromMap=%x != ProofsFromMap root=%x", i, len(m), root, proofRoot)
		}
		if mmRoot := HashFromMap(m); !bytes.Equal(root, mmRoot) {
			t.Fatalf("case %d (n=%d): RootHashFromMap=%x != HashFromMap=%x", i, len(m), root, mmRoot)
		}
	}
}
