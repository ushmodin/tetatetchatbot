package telegram

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MgoDb struct {
	mongo *mgo.Session
	db    string
}

// NewMgoDb create new database object
// con mongodb connection string, ex: host:port
func NewMgoDb(con string, db string) (*MgoDb, error) {
	mongo, err := mgo.Dial(con)
	if err != nil {
		return nil, err
	}
	return &MgoDb{mongo: mongo, db: db}, nil
}

// Close Destroy object
func (db MgoDb) Close() {
	db.mongo.Close()
}

func (db MgoDb) Save(collection string, obj interface{}) error {
	return db.mongo.DB(db.db).C(collection).Insert(obj)
}

func (db MgoDb) IsNotFound(err error) bool {
	return err == mgo.ErrNotFound
}

func (db MgoDb) FindUserByTelegramID(id int) (BotUser, error) {
	var user BotUser
	err := db.mongo.DB(db.db).C("users").Find(bson.M{"TelegramID": id}).One(&user)
	return user, err
}

func (db MgoDb) FindUser(id bson.ObjectId) (BotUser, error) {
	var user BotUser
	err := db.mongo.DB(db.db).C("users").Find(bson.M{"_id": id}).One(&user)
	return user, err
}

func (db MgoDb) SaveUser(user BotUser) error {
	return db.mongo.DB(db.db).C("users").Insert(user)
}

func (db MgoDb) FindDialog(id bson.ObjectId) (Dialog, error) {
	var dialog Dialog
	err := db.mongo.DB(db.db).C("dialogs").Find(bson.M{"_id": id}).One(&dialog)
	return dialog, err
}

func (db MgoDb) DeleteDialog(id bson.ObjectId) error {
	return db.mongo.DB(db.db).C("dialogs").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Status": DIALOG_STATUS_DELETED}})
}

func (db MgoDb) UpdateUserStatus(id bson.ObjectId, status UserStatus) error {
	return db.mongo.DB(db.db).C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Status": status}})
}

func (db MgoDb) UpdateUserPause(id bson.ObjectId, flag bool) error {
	return db.mongo.DB(db.db).C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Pause": flag}})
}

func (db MgoDb) StartDialog(userID bson.ObjectId, chatID int64) error {
	return db.mongo.DB(db.db).C("dialog_requests").Insert(DialogRequest{
		UserID:     userID,
		Processing: false,
		Created:    time.Now().Unix(),
		ChatID:     chatID,
	})
}

func (db MgoDb) BackwardRequestDialog(dlgReq DialogRequest) error {
	dlgReq.Created = time.Now().Unix()
	return db.mongo.DB(db.db).C("dialog_requests").Insert(dlgReq)
}

func (db MgoDb) FindNextDialogRequest() (DialogRequest, error) {
	var req DialogRequest

	_, err := db.mongo.DB(db.db).C("dialog_requests").Find(bson.M{"Processing": false}).Sort("Created").Limit(1).Apply(mgo.Change{
		Update: bson.M{"$set": bson.M{"Processing": true}},
	}, &req)

	return req, err
}

func (db MgoDb) UpdateDialogRequestProcessing(id bson.ObjectId, processing bool) error {
	return db.mongo.DB(db.db).C("dialog_requests").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Processing": processing}})
}

func (db MgoDb) CreateDialog(reqA DialogRequest, reqB DialogRequest) (bson.ObjectId, error) {
	id := bson.NewObjectId()
	dialog := Dialog{
		ID:      id,
		UserA:   reqA.UserID,
		AcceptA: false,
		ChatA:   reqA.ChatID,
		UserB:   reqB.UserID,
		AcceptB: false,
		ChatB:   reqB.ChatID,
		Status:  DIALOG_STATUS_ACTIVE,
	}
	return id, db.mongo.DB(db.db).C("dialogs").Insert(dialog)
}

func (db MgoDb) UpdateUserDialog(userID bson.ObjectId, dialogID *bson.ObjectId) error {
	return db.mongo.DB(db.db).C("users").Update(bson.M{"_id": userID}, bson.M{"$set": bson.M{"DialogID": dialogID}})
}
