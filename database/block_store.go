package database

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

const AfterGenesis = ""

type BlockStore interface {
	Write(blocks ...*Block) (Hash, error)
	Read(after string, limit uint64) ([]Block, error)
	Stream(after string, blockStream chan<- Block)
}

type FileBlockStore struct {
	lock *sync.RWMutex
	file string
}

func NewFileBlockStore(file string) *FileBlockStore {
	return &FileBlockStore{
		lock: &sync.RWMutex{},
		file: file,
	}
}

func (f *FileBlockStore) Write(blocks ...*Block) (Hash, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	file, err := os.OpenFile(f.file, os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return Hash{}, err
	}
	var hash Hash
	for _, block := range blocks {
		hash, err = block.Hash()
		if err != nil {
			return hash, err
		}
		blockFile := BlockFileEntry{
			Hash:  hash,
			Block: block,
		}
		blockFileJson, err := json.Marshal(blockFile)
		if err != nil {
			return hash, err
		}
		if _, err := file.Write(append(blockFileJson, '\n')); err != nil {
			return hash, err
		}
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

	blocks := make([]Block, 0)
	file, err := os.OpenFile(f.file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	if err := f.seek(scanner, after); err != nil {
		return blocks, err
	}
	for i := uint64(0); i < limit && scanner.Scan(); i++ {
		var blockEntry BlockFileEntry
		if err = scanner.Err(); err != nil {
			return blocks, err
		}
		if err = json.Unmarshal(scanner.Bytes(), &blockEntry); err != nil {
			return blocks, err
		}
		blocks = append(blocks, *blockEntry.Block)
	}
	return blocks, nil
}

func (f *FileBlockStore) seek(scanner *bufio.Scanner, after string) error {
	if after == AfterGenesis {
		return nil
	}
	var blockFileObject BlockFileEntry
	for after != blockFileObject.Hash.String() && scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		if err := json.Unmarshal(scanner.Bytes(), &blockFileObject); err != nil {
			return err
		}
	}
	return nil
}
