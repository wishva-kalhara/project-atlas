package data

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
)

type Repository struct {
    filePath string
    mu       sync.Mutex
}

func NewRepository(filePath string) *Repository {
    return &Repository{filePath: filePath}
}

func (r *Repository) Load() (*FbReelsData, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    bytes, err := os.ReadFile(r.filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", r.filePath, err)
    }
    var data FbReelsData
    if err := json.Unmarshal(bytes, &data); err != nil {
        return nil, fmt.Errorf("failed to parse JSON from %s: %w", r.filePath, err)
    }
    return &data, nil
}

func (r *Repository) Save(data *FbReelsData) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    bytes, err := json.MarshalIndent(data, "", "    ")
    if err != nil {
        return fmt.Errorf("failed to marshal data: %w", err)
    }
    tempFile := r.filePath + ".tmp"
    if err := os.WriteFile(tempFile, bytes, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    if err := os.Rename(tempFile, r.filePath); err != nil {
        return fmt.Errorf("failed to atomically commit file: %w", err)
    }
    return nil
}
