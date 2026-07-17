package storage

import (
	"context"
	"fmt"
	"sync"
)

type MemoryStorage struct {
	mu        sync.RWMutex
	documents map[string][]byte
	baseURL   string
}

func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		documents: make(map[string][]byte),
		baseURL:   baseURL,
	}
}

func (s *MemoryStorage) GenerateUploadURL(_ context.Context, verificationID, docType string) (string, error) {
	return fmt.Sprintf("%s/v1/kyc/upload/%s?doc_type=%s", s.baseURL, verificationID, docType), nil
}

func (s *MemoryStorage) GetDocument(_ context.Context, verificationID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.documents[verificationID]
	if !ok {
		return nil, fmt.Errorf("document not found for verification %s", verificationID)
	}
	return data, nil
}

func (s *MemoryStorage) DeleteDocument(_ context.Context, verificationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.documents, verificationID)
	return nil
}

func (s *MemoryStorage) StoreDocument(_ context.Context, verificationID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[verificationID] = data
	return nil
}
