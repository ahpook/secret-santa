package main

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/crypto"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

const (
	documentPrefix = "d:" // Prefix for document entries
	ownerPrefix    = "o:" // Prefix for owner document mapping

	totalSupplyKey = "s" // Key for total supply of documents
)

type DocumentItem struct {
	ID          []byte          // Document ID (hash)
	Name        string          // Document name
	Owner       interop.Hash160 // Owner's address
	ContentHash interop.Hash256 // Hash of the content
}

// AddDocument stores a document with its associated owner in the contract.
func AddDocument(owner interop.Hash160, name string, content []byte) []byte {
	if len(owner) != 20 {
		panic("Invalid owner address")
	}

	ctx := storage.GetContext()
	contentHash := crypto.Sha256(content)

	// Ensure the document doesn't already exist
	if documentExists(ctx, contentHash) {
		panic("Document already exists")
	}

	// Ensure the owner doesn't already own a document with the same name
	if ownerHasDocument(ctx, owner, name) {
		panic("Owner already owns a document with the same name")
	}

	// Create and store the document
	doc := DocumentItem{
		ID:          crypto.Sha256([]byte(name)),
		Name:        name,
		Owner:       owner,
		ContentHash: contentHash,
	}
	setDocument(ctx, contentHash, doc)

	// Map owner to document
	setOwnerDocument(ctx, owner, name, contentHash)

	// Update total supply
	total := getTotalSupply(ctx) + 1
	setTotalSupply(ctx, total)

	runtime.Notify("DocumentAdded", owner, name, contentHash)
	return contentHash
}

// GetDocument retrieves a document by its content hash.
func GetDocument(contentHash []byte) DocumentItem {
	ctx := storage.GetReadOnlyContext()
	return getDocument(ctx, contentHash)
}

// DeleteDocument removes a document and its associated mapping for an owner.
func DeleteDocument(owner interop.Hash160, contentHash interop.Hash256) {
	ctx := storage.GetContext()

	// Ensure the caller is the owner
	doc := getDocument(ctx, contentHash)
	if !runtime.CheckWitness(doc.Owner) {
		panic("Unauthorized")
	}

	// Remove the document
	deleteDocument(ctx, contentHash)

	// Remove the owner's mapping
	deleteOwnerDocument(ctx, owner, doc.Name)

	// Update total supply
	total := getTotalSupply(ctx) - 1
	setTotalSupply(ctx, total)

	runtime.Notify("DocumentDeleted", owner, contentHash)
}

// TotalSupply returns the total number of documents stored in the contract.
func TotalSupply() int {
	ctx := storage.GetReadOnlyContext()
	return getTotalSupply(ctx)
}

// Helper Functions

func documentExists(ctx storage.Context, contentHash []byte) bool {
	key := mkDocumentKey(contentHash)
	return storage.Get(ctx, key) != nil
}

func ownerHasDocument(ctx storage.Context, owner interop.Hash160, name string) bool {
	key := mkOwnerDocumentKey(owner, name)
	return storage.Get(ctx, key) != nil
}

func setDocument(ctx storage.Context, contentHash []byte, doc DocumentItem) {
	key := mkDocumentKey(contentHash)
	val := std.Serialize(doc)
	storage.Put(ctx, key, val)
}

func getDocument(ctx storage.Context, contentHash []byte) DocumentItem {
	key := mkDocumentKey(contentHash)
	val := storage.Get(ctx, key)
	if val == nil {
		panic("Document not found")
	}
	return std.Deserialize(val.([]byte)).(DocumentItem)
}

func deleteDocument(ctx storage.Context, contentHash []byte) {
	key := mkDocumentKey(contentHash)
	storage.Delete(ctx, key)
}

func setOwnerDocument(ctx storage.Context, owner interop.Hash160, name string, contentHash []byte) {
	key := mkOwnerDocumentKey(owner, name)
	storage.Put(ctx, key, contentHash)
}

func deleteOwnerDocument(ctx storage.Context, owner interop.Hash160, name string) {
	key := mkOwnerDocumentKey(owner, name)
	storage.Delete(ctx, key)
}

func getTotalSupply(ctx storage.Context) int {
	val := storage.Get(ctx, []byte(totalSupplyKey))
	if val == nil {
		return 0
	}
	return val.(int)
}

func setTotalSupply(ctx storage.Context, total int) {
	storage.Put(ctx, []byte(totalSupplyKey), total)
}

func mkDocumentKey(contentHash []byte) []byte {
	return append([]byte(documentPrefix), contentHash...)
}

func mkOwnerDocumentKey(owner interop.Hash160, name string) []byte {
	return append([]byte(ownerPrefix+name), owner...)
}
