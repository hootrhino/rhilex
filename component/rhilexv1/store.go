package rhilex

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"bytes"
	"encoding/gob"
)

// EncodeStruct encodes any Go struct into binary form
func EncodeStruct(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

// DecodeStruct decodes binary data into a Go struct pointer
func DecodeStruct(data []byte, out any) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(out)
}

type SqliteCacheStore struct {
	db *sql.DB
}

func NewSqliteCacheStore(path string) (*SqliteCacheStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kv_store (
			slot TEXT NOT NULL,
			key TEXT NOT NULL,
			value BLOB NOT NULL,
			PRIMARY KEY (slot, key)
		);
		CREATE TABLE IF NOT EXISTS stack_store (
			slot TEXT NOT NULL,
			key TEXT NOT NULL,
			idx INTEGER NOT NULL,
			value BLOB NOT NULL,
			PRIMARY KEY (slot, key, idx)
		);
	`)
	if err != nil {
		return nil, err
	}
	return &SqliteCacheStore{db: db}, nil
}

func (s *SqliteCacheStore) Set(slot, key string, value []byte) error {
	_, err := s.db.Exec(`
		INSERT INTO kv_store (slot, key, value)
		VALUES (?, ?, ?)
		ON CONFLICT(slot, key) DO UPDATE SET value = excluded.value
	`, slot, key, value)
	return err
}

func (s *SqliteCacheStore) Get(slot, key string) ([]byte, error) {
	row := s.db.QueryRow(`SELECT value FROM kv_store WHERE slot = ? AND key = ?`, slot, key)
	var value []byte
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return value, err
}

func (s *SqliteCacheStore) Delete(slot, key string) error {
	_, err := s.db.Exec(`DELETE FROM kv_store WHERE slot = ? AND key = ?`, slot, key)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteCacheStore) Exists(slot, key string) (bool, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM kv_store WHERE slot = ? AND key = ?`, slot, key)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (s *SqliteCacheStore) Push(slot, key string, value []byte) error {
	_, err := s.db.Exec(`
		INSERT INTO stack_store (slot, key, idx, value)
		VALUES (?, ?, COALESCE((SELECT MAX(idx)+1 FROM stack_store WHERE slot=? AND key=?), 0), ?)
	`, slot, key, slot, key, value)
	return err
}

func (s *SqliteCacheStore) Pop(slot, key string) ([]byte, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	var idx int
	var value []byte
	err = tx.QueryRow(`
		SELECT idx, value FROM stack_store
		WHERE slot = ? AND key = ?
		ORDER BY idx DESC LIMIT 1
	`, slot, key).Scan(&idx, &value)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM stack_store WHERE slot = ? AND key = ? AND idx = ?`, slot, key, idx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return value, tx.Commit()
}

func (s *SqliteCacheStore) Close() error {
	return s.db.Close()
}
