package telegram

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Db struct {
	mongo *mgo.Session
	db    string
}

// NewDb create new database object
// con mongodb connection string, ex: host:port
func NewDb(con string, db string) (*Db, error) {
	mongo, err := mgo.Dial(con)
	if err != nil {
		return nil, err
	}
	return &Db{mongo: mongo, db: db}, nil
}

// Close Destroy object
func (db Db) Close() {
	db.mongo.Close()
}

func (db Db) Save(collection string, obj interface{}) error {
	return db.mongo.DB(db.db).C(collection).Insert(obj)
}

func (db Db) IsNotFound(err error) bool {
	return err == mgo.ErrNotFound
}

func (db Db) FindUserByTelegramId(id int) (BotUser, error) {
	var user BotUser
	err := db.mongo.DB(db.db).C("users").Find(bson.M{"TelegramID": id}).One(&user)
	return user, err
}

func (db Db) FindUser(id bson.ObjectId) (BotUser, error) {
	var user BotUser
	err := db.mongo.DB(db.db).C("users").Find(bson.M{"_id": id}).One(&user)
	return user, err
}

func (db Db) SaveUser(user BotUser) error {
	return db.mongo.DB(db.db).C("users").Insert(user)
}

func (db Db) FindDialog(id bson.ObjectId) (Dialog, error) {
	var dialog Dialog
	err := db.mongo.DB(db.db).C("dialogs").Find(bson.M{"_id": id}).One(&dialog)
	return dialog, err
}

func (db Db) DeleteDialog(id bson.ObjectId) error {
	return db.mongo.DB(db.db).C("dialogs").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Status": DIALOG_STATUS_DELETED}})
}

func (db Db) UpdateUserStatus(id bson.ObjectId, status UserStatus) error {
	return db.mongo.DB(db.db).C("dialogs").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Status": status}})
}

func (db Db) UpdateUserPause(id bson.ObjectId, flag bool) error {
	return db.mongo.DB(db.db).C("dialogs").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Pause": flag}})
}

func (db Db) StartDialog(userId bson.ObjectId) error {
	return db.mongo.DB(db.db).C("dialog_requests").Insert(DialogRequest{
		UserId:     userId,
		Processing: false,
	})
}

func (db Db) FindNextDialogRequest() (DialogRequest, error) {
	var req DialogRequest

	_, err := db.mongo.DB(db.db).C("dialog_requests").Find(bson.M{"Processing": false}).Apply(mgo.Change{
		Update: bson.M{"$set": bson.M{"Processing": true}},
	}, &req)

	return req, err
}

func (db Db) UpdateDialogRequestProcessing(id bson.ObjectId, processing bool) error {
	return db.mongo.DB(db.db).C("dialog_requests").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"Processing": processing}})
}

func (db Db) CreateDialog(reqA DialogRequest, reqB DialogRequest) (bson.ObjectId, error) {
	id := bson.NewObjectId()
	dialog := Dialog{
		ID:      id,
		UserA:   reqA.UserId,
		AcceptA: false,
		UserB:   reqB.UserId,
		AcceptB: false,
		Status:  DIALOG_STATUS_ACTIVE,
	}
	return id, db.mongo.DB(db.db).C("dialogs").Insert(dialog)
}

func (db Db) UpdateUserDialog(userId bson.ObjectId, dialogId *bson.ObjectId) error {
	return db.mongo.DB(db.db).C("users").Update(bson.M{"_id": userId}, bson.M{"_set": bson.M{"DialogId": dialogId}})
}
