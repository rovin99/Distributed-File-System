
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	ChunkSize   = 1024 * 1024 // 1MB chunks
	StorageDir  = "./storage"
	MetadataDir = "./metadata"
)

// FileMetadata stores information about a split file
type FileMetadata struct {
	FileName    string   `json:"fileName"`
	TotalSize   int64    `json:"totalSize"`
	ChunkHashes []string `json:"chunkHashes"`
	ChunkSizes  []int64  `json:"chunkSizes"`
}

// StorageEngine handles local file operations
type StorageEngine struct {
	basePath     string
	metadataPath string
}

func (se *StorageEngine) readChunk(hash string) (any, any) {
	panic("unimplemented")
}

// NewStorageEngine creates a new storage engine instance
func NewStorageEngine() (*StorageEngine, error) {
	// Create storage directories if they don't exist
	for _, dir := range []string{StorageDir, MetadataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return &StorageEngine{
		basePath:     StorageDir,
		metadataPath: MetadataDir,
	}, nil
}

// SplitFile splits a file into chunks and generates hashes
func (se *StorageEngine) SplitFile(filePath string) (*FileMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	metadata := &FileMetadata{
		FileName:    filepath.Base(filePath),
		TotalSize:   fileInfo.Size(),
		ChunkHashes: make([]string, 0),
		ChunkSizes:  make([]int64, 0),
	}

	buffer := make([]byte, ChunkSize)
	for {
		n, err := file.Read(buffer)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		// Generate hash for chunk
		hash := sha256.Sum256(buffer[:n])
		hashString := hex.EncodeToString(hash[:])
		metadata.ChunkHashes = append(metadata.ChunkHashes, hashString)
		metadata.ChunkSizes = append(metadata.ChunkSizes, int64(n))

		// Store chunk
		if err := se.storeChunk(hashString, buffer[:n]); err != nil {
			return nil, fmt.Errorf("failed to store chunk: %v", err)
		}
	}

	// Store metadata
	if err := se.storeMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %v", err)
	}

	return metadata, nil
}

// storeChunk saves a single chunk to disk
func (se *StorageEngine) storeChunk(hash string, data []byte) error {
	chunkPath := filepath.Join(se.basePath, hash)
	return os.WriteFile(chunkPath, data, 0644)
}

// storeMetadata saves file metadata to disk
func (se *StorageEngine) storeMetadata(metadata *FileMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	metadataPath := filepath.Join(se.metadataPath, metadata.FileName+".json")
	return os.WriteFile(metadataPath, data, 0644)
}

// ReassembleFile reconstructs a file from its chunks
func (se *StorageEngine) ReassembleFile(metadata *FileMetadata, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	for i, hash := range metadata.ChunkHashes {
		chunkPath := filepath.Join(se.basePath, hash)
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk %s: %v", hash, err)
		}

		// Verify chunk size
		if int64(len(chunkData)) != metadata.ChunkSizes[i] {
			return fmt.Errorf("chunk size mismatch for %s", hash)
		}

		// Verify chunk hash
		actualHash := sha256.Sum256(chunkData)
		if hex.EncodeToString(actualHash[:]) != hash {
			return fmt.Errorf("chunk hash mismatch for %s", hash)
		}

		if _, err := outFile.Write(chunkData); err != nil {
			return fmt.Errorf("failed to write chunk to output file: %v", err)
		}
	}

	return nil
}

func (se *StorageEngine) readMetadata(fileName string) (*FileMetadata, error) {
	metadataPath := filepath.Join(se.metadataPath, fileName+".json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metadata FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &metadata, nil
}

func main() {
	engine, err := NewStorageEngine()
	if err != nil {
		fmt.Printf("Failed to create storage engine: %v\n", err)
		return
	}

	// Example usage
	filePath := "example.txt"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", filePath)
		return
	}
	metadata, err := engine.SplitFile(filePath)
	if err != nil {
		fmt.Printf("Failed to split file: %v\n", err)
		return
	}

	outputPath := "reassembled_" + metadata.FileName
	if err := engine.ReassembleFile(metadata, outputPath); err != nil {
		fmt.Printf("Failed to reassemble file: %v\n", err)
		return
	}
}
