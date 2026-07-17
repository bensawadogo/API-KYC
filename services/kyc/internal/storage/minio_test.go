package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildObjectName(t *testing.T) {
	name := buildObjectName("test-id", "passport")
	assert.Contains(t, name, "test-id/passport_")
}

func TestMemoryStorageUpload(t *testing.T) {
	ms := NewMemoryStorage("http://example.com")
	ctx := context.Background()

	url, err := ms.GenerateUploadURL(ctx, "test-123", "nid")
	require.NoError(t, err)
	assert.Contains(t, url, "http://example.com")
	assert.Contains(t, url, "test-123")
}

func TestMemoryStorageGetDocumentNotFound(t *testing.T) {
	ms := NewMemoryStorage("http://example.com")
	ctx := context.Background()

	_, err := ms.GetDocument(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
