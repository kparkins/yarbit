package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func initDataDir(dataDir string) error {
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		return nil
	}
	dbDir := getDatabaseDirectoryPath(dataDir)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return err
	}
	os.Chown(dbDir, os.Getuid(), os.Getuid())
	genesisPath := getGenesisFilePath(dataDir)
	if err := writeGenesisToDisk(genesisPath); err != nil {
		return err
	}
	blockDbPath := getBlockDatabaseFilePath(dataDir)
	if err := writeEmptyBlocksDbToDisk(blockDbPath); err != nil {
		return err
	}
	return nil
}

func getDatabaseDirectoryPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirectoryPath(dataDir), "genesis.json")
}

func getBlockDatabaseFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirectoryPath(dataDir), "block.db")
}

func writeEmptyBlocksDbToDisk(path string) error {
	if err := ioutil.WriteFile(path, []byte(""), os.ModePerm); err != nil {
		return err
	}
	return os.Chown(path, os.Getuid(), os.Getgid())
}

