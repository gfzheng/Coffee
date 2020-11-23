package auth

import (
	"time"

	"github.com/tidwall/buntdb"
)

// NewStore 创建基于buntdb的存储
func NewBuntDBStore(path string) (*BuntDBStore, error) {
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}

	return &BuntDBStore{
		db: db,
	}, nil
}

// Store buntdb存储
type BuntDBStore struct {
	db *buntdb.DB
}

// Set ...
func (a *BuntDBStore) Set(tokenString string, expiration time.Duration) error {
	return a.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(tokenString, "1", &buntdb.SetOptions{Expires: true, TTL: expiration})
		return err
	})
}

// Check ...
func (a *BuntDBStore) Check(tokenString string) (bool, error) {
	var exists bool
	err := a.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(tokenString)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		exists = val == "1"
		return nil
	})
	return exists, err
}

func (a *BuntDBStore) Remove(tokenString string) error {
	return a.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(tokenString)
		return err
	})
}

// Close ...
func (a *BuntDBStore) Close() error {
	return a.db.Close()
}
