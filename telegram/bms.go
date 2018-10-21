package telegram

import (
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

type BoltMessageService struct {
	bolt *bolt.DB
}

type message struct {
	ChatID int64  `json:"chatId"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

func NewBoltMessageService() (*BoltMessageService, error) {
	db, err := bolt.Open("bms.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("messages")); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &BoltMessageService{bolt: db}, nil
}

func (service *BoltMessageService) Close() {
	service.bolt.Close()
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func (service *BoltMessageService) SendServiceMessage(chatId int64, text string) error {
	return service.sendMessage(chatId, "SERVICE", text)
}

func (service *BoltMessageService) SendCompanyMessage(chatId int64, text string) error {
	return service.sendMessage(chatId, "COMPANY", text)
}

func (service *BoltMessageService) sendMessage(chatId int64, typ string, text string) error {
	log.Printf("Send message to chat %d", chatId)
	data, err := json.Marshal(message{
		ChatID: chatId,
		Type:   typ,
		Text:   text,
	})
	if err != nil {
		return err
	}
	return service.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("messages"))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		return b.Put(itob(id), data)
	})
}

func (service *BoltMessageService) Next10() ([]message, error) {
	values := []message{}
	err := service.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("messages"))
		cur := b.Cursor()
		i := 0
		type keyType []byte
		keys := []keyType{}
		for k, v := cur.First(); k != nil && i < 10; k, v = cur.Next() {
			var t message
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			values = append(values, t)
			keys = append(keys, k)
			i++
		}
		for _, key := range keys {
			if err := b.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return values, err
	}
	return values, nil
}
