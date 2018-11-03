package telegram

import (
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

var db *BoltDb
var ms *BoltMessageService
var bot *Bot

func TestBotRun(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

type BotSuite struct {
	suite.Suite
}

func (suite *BotSuite) SetupTest() {
	var err error
	db, err = NewBoltDb()
	if err != nil {
		suite.Fail("Can't create bot", err)
	}
	ms, err = NewBoltMessageService()
	if err != nil {
		suite.Fail("Can't create message service", err)
	}
	bot, err = NewBot(db, ms)
}

func (suite *BotSuite) TearDownTest() {
	db.Close()
	ms.Close()
	os.Remove("bms.db")
	os.Remove("tetatetchatbot.db")
}

func (suite *BotSuite) TestStart() {
	user := User{
		ID:           rand.Int(),
		FirstName:    "Ivan",
		LastName:     "Ivanov",
		UserName:     "IvanovIvan",
		LanguageCode: "RU",
		IsBot:        false,
	}
	chat := Chat{
		ID: rand.Int63(),
	}
	suite.NoError(bot.Start(user, chat), "Can't  start user")
	suite.NoError(bot.Start(user, chat), "Can't  start user")
	messages, err := ms.Next10()
	suite.Nil(err, "Can't get messages")
	suite.Equal(2, len(messages), "Incorrect message count")
	suite.Equal(chat.ID, messages[0].ChatID, "Incorrect message count")
}

func (suite *BotSuite) TestJoinRequests() {
	userA := User{
		ID:           rand.Int(),
		FirstName:    "Ivan",
		LastName:     "Ivanov",
		UserName:     "IvanovIvan",
		LanguageCode: "RU",
		IsBot:        false,
	}
	chatA := Chat{
		ID: rand.Int63(),
	}
	userB := User{
		ID:           rand.Int(),
		FirstName:    "Petr",
		LastName:     "Petrov",
		UserName:     "PetrPetrov",
		LanguageCode: "RU",
		IsBot:        false,
	}
	chatB := Chat{
		ID: rand.Int63(),
	}
	suite.NoError(bot.Start(userA, chatA), "Can't start user")
	suite.NoError(bot.Start(userB, chatB), "Can't start user")

	suite.NoError(bot.Search(userA), "Can't start user")
	suite.NoError(bot.Search(userB), "Can't start user")

	ok, err := bot.JoinRequests()
	suite.NoError(err, "Join error")
	suite.True(ok, "Requests not joined")
	messages, err := ms.Next10()
	suite.NoError(err)
	suite.Equal(4, len(messages), "Incorrect message count")
}
