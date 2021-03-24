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
	Open() error
	Close() error
	Write(block *Block) (Hash, error)
	Read(after string, limit uint64) ([]Block, error)
	Stream(after string, blockStream chan<- Block, group *sync.WaitGroup)
}

type FileBlockStore struct {
	lock   *sync.RWMutex
	file   string
	writer *os.File
}

func NewFileBlockStore(file string) *FileBlockStore {
	return &FileBlockStore{
		lock:   &sync.RWMutex{},
		writer: nil,
		file:   file,
	}
}

func (f *FileBlockStore) Open() error {
	var e error
	f.writer, e = os.OpenFile(f.file, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	return e
}

func (f *FileBlockStore) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	return nil
}

func (f *FileBlockStore) Write(block *Block) (Hash, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	hash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}
	blockFile := BlockFs{
		Key:   hash,
		Value: *block,
	}
	blockFileJson, err := json.Marshal(blockFile)
	if err != nil {
		return Hash{}, err
	}
	fmt.Printf("Persisting new block to disk: \n")
	fmt.Printf("\t%x\n", hash)
	if _, err := f.writer.Write(append(blockFileJson, '\n')); err != nil {
		return Hash{}, err
	}
	return hash, nil
}

func (f *FileBlockStore) Stream(after string, blockStream chan<- Block, group *sync.WaitGroup) {
	group.Add(1)
	go func() {
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
		group.Done()
	}()
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
