package storage

import (
    "github.com/boltdb/bolt"
)

type LocalStorage struct {
    db *bolt.DB
}

func NewLocalStorage(path string) (*LocalStorage, error) {
    db, err := bolt.Open(path, 0600, nil)
    if err != nil {
        return nil, err
    }

    return &LocalStorage{db: db}, nil
}

func (ls *LocalStorage) StoreChunk(chunkHash string, data []byte) error {
    return ls.db.Update(func(tx *bolt.Tx) error {
        bucket, err := tx.CreateBucketIfNotExists([]byte("chunks"))
        if err != nil {
            return err
        }
        return bucket.Put([]byte(chunkHash), data)
    })
}
