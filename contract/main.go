package main

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
	// Other needed packages from neo-go’s standard library
)

// SecretSantaContract is our contract struct; it's optional if you want
// to keep everything in top-level functions, but can help organize code.
type SecretSantaContract struct{}

// Constants or other top-level variables
const (
	participantPrefix = "participant:"
	pairsKey          = "pairs"
)

// init or _deploy can be used to perform contract initialization
// if necessary. We keep it empty if not needed.
func init() {
	// Called when contract is being loaded.
	// (You might also see _deploy(bool, []byte) pattern in other examples.)
}

// AddParticipant stores a new participant’s address (or public key, etc.)
func AddParticipant(name string) bool {
	ctx := storage.GetContext()
	// Example: store participants with a prefix
	fullKey := participantPrefix + name

	existing := storage.Get(ctx, []byte(fullKey))
	if len(existing) != 0 {
		// Participant already exists
		return false
	}

	storage.Put(ctx, []byte(fullKey), []byte("registered"))
	return true
}

// GetAllParticipants returns an array of participant names from storage
// This is a simplistic approach and might not scale well if participants are numerous
func GetAllParticipants() []string {
	ctx := storage.GetContext()
	var result []string

	iterator := storage.Find(ctx, []byte(participantPrefix), storage.None)
	for iterator.Next() {
		key := iterator.Key()
		// Key is something like "participant:John"
		// Convert key bytes to string, then remove prefix
		keyStr := string(key)
		name := keyStr[len(participantPrefix):]
		result = append(result, name)
	}
	return result
}

// ShuffleAndStorePairs is a naive pseudo-random pairing function.
// In production, you'd likely need a more secure random source.
func ShuffleAndStorePairs() bool {
	participants := GetAllParticipants()
	if len(participants) < 2 {
		// Not enough participants to pair
		return false
	}

	// Pseudo-random: extremely naive approach using the current block timestamp
	// **NOT** secure for real usage. Just for demonstration.
	seed := int(runtime.GetTime())
	shuffled := simpleShuffle(participants, seed)

	// We will store the mapping as a single JSON string for example
	pairsMap := make(map[string]string)
	for i, p := range participants {
		// The i-th participant is assigned to the i-th in the shuffled list.
		pairsMap[p] = shuffled[i]
	}

	encoded := marshalPairs(pairsMap)
	ctx := storage.GetContext()
	storage.Put(ctx, []byte(pairsKey), encoded)
	return true
}

// GetPairedWith returns the name that `participantName` is assigned to.
func GetPairedWith(participantName string) string {
	ctx := storage.GetContext()
	encoded := storage.Get(ctx, []byte(pairsKey))
	if len(encoded) == 0 {
		return ""
	}
	pairsMap := unmarshalPairs(encoded)
	return pairsMap[participantName]
}

// Below are helper functions to keep the example short.
// Because of language limitations in neo-go, you often must rely on simple
// data structures, and you can’t use the standard "encoding/json" package in the same way
// as normal Go. For a real scenario, you might store pairs as a simple list of (key->value) in storage
// or use neo-go’s serialized struct approach.

func simpleShuffle(arr []string, seed int) []string {
	// Very naive shuffle for demonstration only
	shuffled := make([]string, len(arr))
	copy(shuffled, arr)
	for i := range shuffled {
		j := (i + seed) % len(shuffled)
		// swap
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

// Note: For real usage, handle serialization carefully
// with built-in neo-go serialization methods (like JSON* or
// binary serialization from the interop packages).
// The example below is purely illustrative.

func marshalPairs(pm map[string]string) []byte {
	// Very naive "key1:value1;key2:value2" style encoding
	str := ""
	for k, v := range pm {
		str += k + ":" + v + ";"
	}
	return []byte(str)
}

func unmarshalPairs(data []byte) map[string]string {
	pm := make(map[string]string)
	str := string(data)
	if len(str) == 0 {
		return pm
	}
	pairs := split(str, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := split(pair, ":")
		if len(kv) == 2 {
			pm[kv[0]] = kv[1]
		}
	}
	return pm
}

// split is a trivial string-split function that doesn't rely on reflection
// or other advanced Go features.
func split(s, sep string) []string {
	var result []string
	start := 0
	for {
		idx := indexOf(s, sep, start)
		if idx < 0 {
			result = append(result, s[start:])
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}
	return result
}

// indexOf finds the first instance of sep in s starting at offset start.
func indexOf(s, sep string, start int) int {
	for i := start; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// Main entrypoint for the contract: This can be a typical "switch" on
// operation names if you prefer standard neo-go contract patterns.
func Main(operation string, args []interface{}) interface{} {
	switch operation {
	case "addParticipant":
		if len(args) < 1 {
			return false
		}
		name := args[0].(string)
		return AddParticipant(name)
	case "getAllParticipants":
		return GetAllParticipants()
	case "shuffleAndStorePairs":
		return ShuffleAndStorePairs()
	case "getPairedWith":
		if len(args) < 1 {
			return ""
		}
		name := args[0].(string)
		return GetPairedWith(name)
	default:
		runtime.Log("Unknown operation: " + operation)
		return false
	}
}
