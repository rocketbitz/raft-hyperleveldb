package rafthyperleveldb

import (
	"os"
	"testing"

	"github.com/hashicorp/raft/bench"
)

func BenchmarkHyperLevelDBStore_FirstIndex(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.FirstIndex(b, store)
}

func BenchmarkHyperLevelDBStore_LastIndex(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.LastIndex(b, store)
}

func BenchmarkHyperLevelDBStore_GetLog(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.GetLog(b, store)
}

func BenchmarkHyperLevelDBStore_StoreLog(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.StoreLog(b, store)
}

func BenchmarkHyperLevelDBStore_StoreLogs(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.StoreLogs(b, store)
}

func BenchmarkHyperLevelDBStore_DeleteRange(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.DeleteRange(b, store)
}

func BenchmarkHyperLevelDBStore_Set(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.Set(b, store)
}

func BenchmarkHyperLevelDBStore_Get(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.Get(b, store)
}

func BenchmarkHyperLevelDBStore_SetUint64(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.SetUint64(b, store)
}

func BenchmarkHyperLevelDBStore_GetUint64(b *testing.B) {
	store := testHyperLevelDBStore(b)
	defer store.Close()
	defer os.Remove(store.path)

	raftbench.GetUint64(b, store)
}
