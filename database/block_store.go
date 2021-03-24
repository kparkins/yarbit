package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const AfterNone = ""

type BlockStore interface {
	Write(block *Block) (Hash, error)
	Read(after string, limit uint64) ([]Block, error)
	Stream(after string, blockStream chan<- Block)
}

type FileBlockStore struct {
	lock   *sync.RWMutex
	file   string
}

func NewFileBlockStore(file string) *FileBlockStore {
	return &FileBlockStore{
		lock:   &sync.RWMutex{},
		file:   file,
	}
}

func (f *FileBlockStore) Write(block *Block) (Hash, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	hash, err := block.Hash()
	if err != nil {
		return hash, err
	}
	blockFile := BlockFs{
		Key:   hash,
		Value: *block,
	}
	blockFileJson, err := json.Marshal(blockFile)
	if err != nil {
		return hash, err
	}
	file, err := os.Open(f.file)
	if err != nil {
		return hash, err
	}
	fmt.Printf("Persisting new block to disk: \n")
	fmt.Printf("\t%x\n", hash)
	if _, err := file.Write(append(blockFileJson, '\n')); err != nil {
		return hash, err
	}
	return hash, nil
}

func (f *FileBlockStore) Stream(after string, blockStream chan<- Block) {
	batch := uint64(cap(blockStream))
	for {
		blocks, err := f.Read(after, batch)
		if err != nil || len(blocks) < 1 {
			goto done
		}
		for _, b := range blocks {
			blockStream <- b
		}
		h, err := blocks[len(blocks)-1].Hash()
		if err != nil || uint64(len(blocks)) < batch {
			goto done
		}
		after = h.String()
	}
done:
	close(blockStream)
}

func (f *FileBlockStore) Read(after string, limit uint64) ([]Block, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	var blocks []Block
	file, err := os.OpenFile(f.file, os.O_APPEND|os.O_RDONLY, os.ModePerm)
	if err != nil {
		return blocks, nil
	}
	var blockFileObject BlockFs
	scanner := bufio.NewScanner(file)
	for after != "" && scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(scanner.Bytes(), &blockFileObject); err != nil {
			return nil, err
		}
		cursor := blockFileObject.Key.String()
		if cursor == after {
			break
		}
	}
	for i := uint64(0); i < limit && scanner.Scan(); i++ {
		if err = scanner.Err(); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(scanner.Bytes(), &blockFileObject); err != nil {
			return nil, err
		}
		blocks = append(blocks, blockFileObject.Value)
	}
	return blocks, nil
}
