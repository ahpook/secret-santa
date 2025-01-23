package main

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/crypto"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

const (
	documentPrefix = "d:"
	ownerPrefix    = "o:"

	totalSupplyKey = "s"
)

type DocumentItem struct {
	ID          []byte
	Name        string
	Owner       interop.Hash160
	ContentHash interop.Hash256
}

func AddDocument(owner interop.Hash160, name string, content []byte) []byte {
	if len(owner) != 20 {
		panic("Invalid owner address")
	}

	ctx := storage.GetContext()
	contentHash := crypto.Sha256(content)

	if documentExists(ctx, contentHash) {
		panic("Document already exists")
	}

	if ownerHasDocument(ctx, owner, name) {
		panic("Owner already owns a document with the same name")
	}

	doc := DocumentItem{
		ID:          crypto.Sha256([]byte(name)),
		Name:        name,
		Owner:       owner,
		ContentHash: contentHash,
	}
	setDocument(ctx, contentHash, doc)

	setOwnerDocument(ctx, owner, name, contentHash)

	total := getTotalSupply(ctx) + 1
	setTotalSupply(ctx, total)

	runtime.Notify("DocumentAdded", owner, name, contentHash)
	return contentHash
}

func GetDocument(contentHash []byte) DocumentItem {
	ctx := storage.GetReadOnlyContext()
	return getDocument(ctx, contentHash)
}

func DeleteDocument(owner interop.Hash160, contentHash interop.Hash256) {
	ctx := storage.GetContext()

	doc := getDocument(ctx, contentHash)
	if !runtime.CheckWitness(doc.Owner) {
		panic("Unauthorized")
	}

	deleteDocument(ctx, contentHash)

	deleteOwnerDocument(ctx, owner, doc.Name)

	total := getTotalSupply(ctx) - 1
	setTotalSupply(ctx, total)

	runtime.Notify("DocumentDeleted", owner, contentHash)
}

func TotalSupply() int {
	ctx := storage.GetReadOnlyContext()
	return getTotalSupply(ctx)
}

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
