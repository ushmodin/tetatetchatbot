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

func (db Db) FindUser(id int) (BotUser, error) {
	var user BotUser
	err := db.mongo.DB(db.db).C("users").Find(bson.M{"telegramID": id}).One(&user)
	return user, err
}

func (db Db) SaveUser(user BotUser) error {
	return db.mongo.DB(db.db).C("users").Insert(user)
}
