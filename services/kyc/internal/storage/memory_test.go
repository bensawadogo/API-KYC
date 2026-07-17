package storage

import (
	"context"
	"testing"
)

func TestMemoryStorage_GenerateUploadURL(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	url, err := s.GenerateUploadURL(context.Background(), "v123", "NATIONAL_ID")
	if err != nil {
		t.Fatal(err)
	}
	expected := "http://localhost:8080/v1/kyc/upload/v123?doc_type=NATIONAL_ID"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestMemoryStorage_StoreAndGetDocument(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	docData := []byte("fake-document-image-data")

	err := s.StoreDocument(context.Background(), "v1", docData)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := s.GetDocument(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if string(retrieved) != string(docData) {
		t.Errorf("expected %s, got %s", string(docData), string(retrieved))
	}
}

func TestMemoryStorage_GetDocument_NotFound(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	_, err := s.GetDocument(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing document")
	}
}

func TestMemoryStorage_DeleteDocument(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	s.StoreDocument(context.Background(), "v2", []byte("data"))

	err := s.DeleteDocument(context.Background(), "v2")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.GetDocument(context.Background(), "v2")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestMemoryStorage_DeleteDocument_NonExistent(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	err := s.DeleteDocument(context.Background(), "v3")
	if err != nil {
		t.Fatal("expected no error deleting non-existent")
	}
}

func TestMemoryStorage_StoreAndDeleteMultiple(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	doc1 := []byte("doc1")
	doc2 := []byte("doc2")

	s.StoreDocument(context.Background(), "v1", doc1)
	s.StoreDocument(context.Background(), "v2", doc2)

	s.DeleteDocument(context.Background(), "v1")

	d1, err := s.GetDocument(context.Background(), "v1")
	if err == nil {
		t.Fatal("expected error, got data")
	}
	_ = d1

	d2, err := s.GetDocument(context.Background(), "v2")
	if err != nil {
		t.Fatal(err)
	}
	if string(d2) != string(doc2) {
		t.Errorf("expected doc2, got %s", string(d2))
	}
}

func TestMemoryStorage_StoreOverwrite(t *testing.T) {
	s := NewMemoryStorage("http://localhost:8080")
	s.StoreDocument(context.Background(), "v1", []byte("original"))
	s.StoreDocument(context.Background(), "v1", []byte("updated"))

	doc, err := s.GetDocument(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if string(doc) != "updated" {
		t.Errorf("expected updated, got %s", string(doc))
	}
}
