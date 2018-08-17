package rafthyperleveldb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/hashicorp/raft"
	"github.com/jmhodges/levigo"
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

var (
	// Bucket names we perform transactions in
	dbLogs = []byte("logs")
	dbConf = []byte("conf")

	// An error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
)

var errInvalidLog = errors.New("invalid log")

// HyperLevelDBStore provides access to BoltDB for Raft to store and retrieve
// log entries. It also provides key/value storage, and can be used as
// a LogStore and StableStore.
type HyperLevelDBStore struct {
	// db is the underlying handle to the db.
	db *levigo.DB

	// The path to the Bolt database file
	path string

	wo *levigo.WriteOptions
	ro *levigo.ReadOptions
}

type Level int

const (
	Low    Level = -1
	Medium Level = 0
	High   Level = 1
)

// NewHyperLevelDBStore takes a file path and returns a connected Raft backend.
func NewHyperLevelDBStore(path string, durability Level) (*HyperLevelDBStore, error) {
	var (
		opts = levigo.NewOptions()
		wo   = levigo.NewWriteOptions()
		ro   = levigo.NewReadOptions()
	)

	if durability <= Low {
		wo.SetSync(false)
	}

	opts.SetCreateIfMissing(true)

	// Try to connect
	db, err := levigo.Open(path, opts)
	if err != nil {
		return nil, err
	}

	// Create the new store
	store := &HyperLevelDBStore{
		db:   db,
		path: path,
		wo:   wo,
		ro:   ro,
	}

	return store, nil
}

// Close is used to gracefully close the DB connection.
func (b *HyperLevelDBStore) Close() error {
	b.db.Close()
	return nil
}

// FirstIndex returns the first known index from the Raft log.
func (b *HyperLevelDBStore) FirstIndex() (uint64, error) {
	var n uint64
	iter := b.db.NewIterator(b.ro)
	for {
		iter.Seek(dbLogs)
		if !iter.Valid() {
			break
		}

		key := iter.Key()
		if !bytes.HasPrefix(key, dbLogs) {
			break
		}
		n = bytesToUint64(key[len(dbLogs):])
		break
	}

	return n, nil
}

// LastIndex returns the last known index from the Raft log.
func (b *HyperLevelDBStore) LastIndex() (uint64, error) {
	var n uint64
	iter := b.db.NewIterator(b.ro)
	for {
		iter.SeekToLast()
		if !iter.Valid() {
			break
		}

		key := iter.Key()
		if !bytes.HasPrefix(key, dbLogs) {
			break
		}
		n = bytesToUint64(key[len(dbLogs):])
		break
	}

	return n, nil
}

// GetLog is used to retrieve a log from BoltDB at a given index.
func (b *HyperLevelDBStore) GetLog(idx uint64, log *raft.Log) error {
	key := append(dbLogs, uint64ToBytes(idx)...)
	val, err := b.db.Get(b.ro, key)
	if err != nil {
		return err
	}
	if val == nil {
		return raft.ErrLogNotFound
	}
	return decodeLog(val, log)
}

// StoreLog is used to store a single raft log
func (b *HyperLevelDBStore) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// StoreLogs is used to store a set of raft logs
func (b *HyperLevelDBStore) StoreLogs(logs []*raft.Log) error {
	batch := levigo.NewWriteBatch()
	for _, log := range logs {
		key := append(dbLogs, uint64ToBytes(log.Index)...)
		batch.Put(key, encodeLog(log))
	}
	return b.db.Write(b.wo, batch)
}

// DeleteRange is used to delete logs within a given range inclusively.
func (b *HyperLevelDBStore) DeleteRange(min, max uint64) error {
	batch := levigo.NewWriteBatch()
	prefix := append(dbLogs, uint64ToBytes(min)...)
	iter := b.db.NewIterator(b.ro)
	iter.Seek(prefix)
	for {
		if !iter.Valid() {
			break
		}

		key := iter.Key()
		if !bytes.HasPrefix(key, dbLogs) {
			break
		}
		if bytesToUint64(key[len(dbLogs):]) > max {
			break
		}
		batch.Delete(key)
		iter.Next()
	}
	return b.db.Write(b.wo, batch)
}

// Set is used to set a key/value set outside of the raft log
func (b *HyperLevelDBStore) Set(k, v []byte) error {
	return b.db.Put(b.wo, append(dbConf, k...), v)
}

// Get is used to retrieve a value from the k/v store by key
func (b *HyperLevelDBStore) Get(k []byte) ([]byte, error) {
	val, err := b.db.Get(b.ro, append(dbConf, k...))
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, ErrKeyNotFound
	}
	return bcopy(val), nil
}

// SetUint64 is like Set, but handles uint64 values
func (b *HyperLevelDBStore) SetUint64(key []byte, val uint64) error {
	return b.Set(key, uint64ToBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (b *HyperLevelDBStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

// Peers returns raft peers
func (b *HyperLevelDBStore) Peers() ([]string, error) {
	var peers []string
	val, err := b.Get([]byte("peers"))
	if err != nil {
		if err == ErrKeyNotFound {
			return []string{}, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(val, &peers); err != nil {
		return nil, err
	}
	return peers, nil
}

// SetPeers sets raft peers
func (b *HyperLevelDBStore) SetPeers(peers []string) error {
	data, err := json.Marshal(peers)
	if err != nil {
		return err
	}
	return b.Set([]byte("peers"), data)
}

func bcopy(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)
	return r
}

// Decode reverses the encode operation on a byte slice input
func decodeLog(buf []byte, log *raft.Log) error {
	if len(buf) < 25 {
		return errInvalidLog
	}
	log.Index = binary.LittleEndian.Uint64(buf[0:8])
	log.Term = binary.LittleEndian.Uint64(buf[8:16])
	log.Type = raft.LogType(buf[16])
	log.Data = make([]byte, binary.LittleEndian.Uint64(buf[17:25]))
	if len(buf[25:]) < len(log.Data) {
		return errInvalidLog
	}
	copy(log.Data, buf[25:])
	return nil
}

// Encode writes an encoded object to a new bytes buffer
func encodeLog(log *raft.Log) []byte {
	var buf []byte
	var num = make([]byte, 8)
	binary.LittleEndian.PutUint64(num, log.Index)
	buf = append(buf, num...)
	binary.LittleEndian.PutUint64(num, log.Term)
	buf = append(buf, num...)
	buf = append(buf, byte(log.Type))
	binary.LittleEndian.PutUint64(num, uint64(len(log.Data)))
	buf = append(buf, num...)
	buf = append(buf, log.Data...)
	return buf
}

// Converts bytes to an integer
func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Converts a uint to a byte slice
func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
