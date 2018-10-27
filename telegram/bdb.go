package telegram

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
)

type BoltDb struct {
	bolt *bolt.DB
}

var (
	NotFoundError = errors.New("#Notfound")
)

func NewBoltDb() (*BoltDb, error) {
	db, err := bolt.Open("tetatetchatbot.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte("telegram_user_idx")); err != nil {
			return err
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("users")); err != nil {
			return err
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("dialogs")); err != nil {
			return err
		}
		return nil
	})
	return &BoltDb{bolt: db}, nil
}

func (db *BoltDb) Close() {
	db.bolt.Close()
}

func (db *BoltDb) FindUserByTelegramID(id int) (BotUser, error) {
	var userID []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("telegram_user_idx"))
		userID = b.Get([]byte(strconv.Itoa(id)))
		return nil
	})
	if err != nil {
		return BotUser{}, err
	}
	if userID == nil {
		return BotUser{}, NotFoundError
	}
	var user []byte
	err = db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		user = b.Get(userID)
		return nil
	})
	if user == nil {
		return BotUser{}, NotFoundError
	}
	var botUser BotUser
	json.Unmarshal(user, &botUser)
	return botUser, nil
}

func (db *BoltDb) IsNotFound(err error) bool {
	return err == NotFoundError
}

func (db *BoltDb) SaveUser(user BotUser) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if err := b.Put([]byte(user.ID), data); err != nil {
			return err
		}
		b = tx.Bucket([]byte("telegram_user_idx"))
		if err = b.Put([]byte(strconv.Itoa(user.TelegramID)), []byte(user.ID)); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (db *BoltDb) FindUser(id bson.ObjectId) (BotUser, error) {
	var data []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		data = b.Get([]byte(id))
		return nil
	})
	if err != nil {
		return BotUser{}, err
	}
	if data == nil {
		return BotUser{}, NotFoundError
	}
	var botUser BotUser
	err = json.Unmarshal(data, &botUser)
	if err != nil {
		return BotUser{}, err
	}
	return botUser, nil
}

func (db *BoltDb) FindDialog(id bson.ObjectId) (Dialog, error) {
	var data []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("dialogs"))
		data = b.Get([]byte(id))
		return nil
	})
	if err != nil {
		return Dialog{}, err
	}
	if data == nil {
		return Dialog{}, NotFoundError
	}
	var dialog Dialog
	err = json.Unmarshal(data, &dialog)
	if err != nil {
		return Dialog{}, err
	}
	return dialog, nil
}

func (db *BoltDb) UpdateUserDialog(userID bson.ObjectId, dialogID *bson.ObjectId) error {
	var data []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		data = b.Get([]byte(userID))
		return nil
	})
	if err != nil {
		return err
	}
	if data == nil {
		return NotFoundError
	}
	var botUser BotUser
	if err = json.Unmarshal(data, &botUser); err != nil {
		return err
	}
	botUser.DialogID = dialogID
	if data, err = json.Marshal(botUser); err != nil {
		return err
	}
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(userID), data)
	})
	return nil
}
func (db *BoltDb) UpdateUserPause(userID bson.ObjectId, flag bool) error {
	var data []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		data = b.Get([]byte(userID))
		return nil
	})
	if err != nil {
		return err
	}
	if data == nil {
		return NotFoundError
	}
	var botUser BotUser
	if err = json.Unmarshal(data, &botUser); err != nil {
		return err
	}
	botUser.Pause = flag
	if data, err = json.Marshal(botUser); err != nil {
		return err
	}
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(userID), data)
	})
	return nil
}

func (db *BoltDb) UpdateUserStatus(userID bson.ObjectId, status UserStatus) error {
	var data []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		data = b.Get([]byte(userID))
		return nil
	})
	if err != nil {
		return err
	}
	if data == nil {
		return NotFoundError
	}
	var botUser BotUser
	if err = json.Unmarshal(data, &botUser); err != nil {
		return err
	}
	botUser.Status = status
	if data, err = json.Marshal(botUser); err != nil {
		return err
	}
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(userID), data)
	})
	return nil
}

func (db *BoltDb) CreateDialog(reqA DialogRequest, reqB DialogRequest) (bson.ObjectId, error) {
	return "", errors.New("Not implemeted yet")
}

func (db *BoltDb) StartDialog(userID bson.ObjectId) error {
	return errors.New("Not implemeted yet")
}

func (db *BoltDb) DeleteDialog(id bson.ObjectId) error {
	return errors.New("Not implemeted yet")
}

func (db *BoltDb) FindNextDialogRequest() (DialogRequest, error) {
	return DialogRequest{}, errors.New("Not implemeted yet")
}

func (db *BoltDb) BackwardRequestDialog(dlgReq DialogRequest) error {
	return errors.New("Not implemeted yet")
}

func (db *BoltDb) UpdateDialogRequestProcessing(id bson.ObjectId, processing bool) error {
	return errors.New("Not implemeted yet")
}
