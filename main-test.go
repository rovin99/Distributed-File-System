package main

import (
    "bytes"
    "crypto/rand"
    "os"
    "path/filepath"
    "testing"
)

func TestStorageEngine(t *testing.T) {
    // Create a temporary test file
    testData := make([]byte, 3*ChunkSize+1024) // Create data that will split into multiple chunks
    _, err := rand.Read(testData)
    if err != nil {
        t.Fatalf("Failed to generate test data: %v", err)
    }

    testFile := filepath.Join(os.TempDir(), "test_file.bin")
    if err := os.WriteFile(testFile, testData, 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    defer os.Remove(testFile)

    // Initialize storage engine
    engine, err := NewStorageEngine()
    if err != nil {
        t.Fatalf("Failed to create storage engine: %v", err)
    }

    // Test file splitting
    metadata, err := engine.SplitFile(testFile)
    if err != nil {
        t.Fatalf("Failed to split file: %v", err)
    }

    // Verify number of chunks
    expectedChunks := (len(testData) + ChunkSize - 1) / ChunkSize
    if len(metadata.ChunkHashes) != expectedChunks {
        t.Errorf("Expected %d chunks, got %d", expectedChunks, len(metadata.ChunkHashes))
    }

    // Test file reassembly
    outputFile := filepath.Join(os.TempDir(), "reassembled_test_file.bin")
    if err := engine.ReassembleFile(metadata, outputFile); err != nil {
        t.Fatalf("Failed to reassemble file: %v", err)
    }
    defer os.Remove(outputFile)

    // Verify reassembled content
    reassembledData, err := os.ReadFile(outputFile)
    if err != nil {
        t.Fatalf("Failed to read reassembled file: %v", err)
    }

    if !bytes.Equal(testData, reassembledData) {
        t.Error("Reassembled file content doesn't match original")
    }
}

func TestChunkSizeValidation(t *testing.T) {
    // Create a small test file
    testData := []byte("Hello, World!")
    testFile := filepath.Join(os.TempDir(), "small_test.txt")
    if err := os.WriteFile(testFile, testData, 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    defer os.Remove(testFile)

    engine, err := NewStorageEngine()
    if err != nil {
        t.Fatalf("Failed to create storage engine: %v", err)
    }

    metadata, err := engine.SplitFile(testFile)
    if err != nil {
        t.Fatalf("Failed to split file: %v", err)
    }

    if len(metadata.ChunkHashes) != 1 {
        t.Errorf("Expected 1 chunk for small file, got %d", len(metadata.ChunkHashes))
    }

    if metadata.TotalSize != int64(len(testData)) {
        t.Errorf("Expected total size %d, got %d", len(testData), metadata.TotalSize)
    }
}