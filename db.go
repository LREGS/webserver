package main

import (
	"os"
	"sync"
)

type Chirp struct {
	id   string
	body string
}

type DB struct {
	path string
	//means that the connection needs to be locked and unlocked
	mux *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(path)
			if err != nil {
				return nil, err
			}
			defer file.Close()
		} else {
			return nil, err
		}
	}

	return &DB{
		//returning a addres to this struct that provides a "connection" (path) to a db and a Rmutex - whihc allows us to lock read and write
		// this is where you can have multiple readers but lock to a single writer
		path: path,
		mux:  &sync.RWMutex{},
	}, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error)

func (db *DB) ensureDB() error
