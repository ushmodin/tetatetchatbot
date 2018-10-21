package telegram

import (
	"encoding/binary"
	"encoding/json"

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
