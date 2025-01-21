package client

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

type Pairing struct {
	Giver    interop.PublicKey
	Receiver interop.PublicKey
}

const playersListKey = "playersList"
const pairingsKey = "pairings"

// Регистрация нового игрока
func RegisterPlayer(pubKey interop.PublicKey) {
	ctx := storage.GetContext()

	// Проверяем, зарегистрирован ли игрок
	if storage.Get(ctx, pubKey) != nil {
		panic("Player already registered")
	}

	// Добавляем публичный ключ игрока в хранилище
	storage.Put(ctx, pubKey, true)

	// Получаем список игроков
	playersListBytes := storage.Get(ctx, playersListKey)
	var playersList []interop.PublicKey
	if playersListBytes != nil {
		playersList = std.Deserialize(playersListBytes.([]byte)).([]interop.PublicKey)
	}

	// Добавляем игрока в список
	playersList = append(playersList, pubKey)

	// Сохраняем обновленный список игроков
	storage.Put(ctx, playersListKey, std.Serialize(playersList))

	runtime.Log("Player registered")
}

// Генерация пар и шифрование сообщений
func GeneratePairings() {
	ctx := storage.GetContext()

	// Получаем список всех игроков
	playersListBytes := storage.Get(ctx, playersListKey)
	if playersListBytes == nil {
		panic("No players registered")
	}

	playersList := std.Deserialize(playersListBytes.([]byte)).([]interop.PublicKey)
	n := len(playersList)
	if n < 2 {
		panic("Not enough players to generate pairings")
	}

	// Генерация пар
	pairings := make([]Pairing, n)
	shuffledPlayers := shuffle(playersList)
	for i := 0; i < n; i++ {
		pairings[i] = Pairing{
			Giver:    shuffledPlayers[i],
			Receiver: shuffledPlayers[(i+1)%n],
		}
	}

	// Сохранение сообщений в зашифрованной форме (имитация)
	for _, pairing := range pairings {
		message := "You are gifting to: " + std.Base64Encode(pairing.Receiver)
		storage.Put(ctx, std.Base64Encode(pairing.Giver), message)
	}

	runtime.Log("Pairings generated and messages encrypted")
}

// Метод получения сообщения (расшифровка выполняется клиентом)
func RevealRecipient(pubKey interop.PublicKey) string {
	ctx := storage.GetReadOnlyContext()
	encryptedMessage := storage.Get(ctx, std.Base64Encode(pubKey))
	if encryptedMessage == nil {
		panic("No message found for this public key")
	}

	return string(encryptedMessage.([]byte))
}

// Метод получения всех игроков
func GetAllPlayers() []interop.PublicKey {
	ctx := storage.GetReadOnlyContext()
	playersListBytes := storage.Get(ctx, playersListKey)
	var playersList []interop.PublicKey
	if playersListBytes != nil {
		playersList = std.Deserialize(playersListBytes.([]byte)).([]interop.PublicKey)
	}
	return playersList
}

// Вспомогательная функция для перемешивания списка игроков
// shuffle перемешивает массив публичных ключей
func shuffle(players []interop.PublicKey) []interop.PublicKey {
	n := len(players)
	shuffled := make([]interop.PublicKey, n)
	for i, player := range players {
		shuffled[i] = player // Копируем элементы вручную
	}

	for i := n - 1; i > 0; i-- {
		j := runtime.GetRandom() % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

func DeleteAllPlayers() {
	ctx := storage.GetContext()

	// Получаем список всех игроков
	playersListBytes := storage.Get(ctx, "playersList")
	if playersListBytes == nil {
		runtime.Log("No players to delete")
		return // Если список пуст, просто выходим
	}

	// Десериализуем список игроков
	var playersList []interop.PublicKey
	playersList = std.Deserialize(playersListBytes.([]byte)).([]interop.PublicKey)

	// Удаляем каждого игрока из хранилища
	for _, pubKey := range playersList {
		storage.Delete(ctx, pubKey) // Удаляем игрока по его публичному ключу
	}

	// Очищаем список игроков
	storage.Delete(ctx, "playersList") // Удаляем список игроков

	// Логирование
	runtime.Log("All players deleted")
	runtime.Notify("All players deletedd", 2)
}
