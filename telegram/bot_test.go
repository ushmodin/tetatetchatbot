package telegram

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestBotRun(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

type BotSuite struct {
	suite.Suite
	db  *MgoDb
	ms  *BoltMessageService
	bot *Bot
}

func (suite *BotSuite) SetupTest() {
	var err error
	suite.db, err = NewMgoDb("localhost", "testtetatet")
	require.NoError(suite.T(), err)
	suite.ms, err = NewBoltMessageService()
	require.NoError(suite.T(), err)
	suite.bot, err = NewBot(suite.db, suite.ms)
	require.NoError(suite.T(), err)
}

func (suite *BotSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.ms != nil {
		suite.ms.Close()
	}
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
	require.NoError(suite.T(), suite.bot.Start(user, chat), "Can't  start user")
	require.NoError(suite.T(), suite.bot.Start(user, chat), "Can't  start user")
	messages, err := suite.ms.Next10()
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
	require.NoError(suite.T(), suite.bot.Start(userA, chatA), "Can't start user")
	require.NoError(suite.T(), suite.bot.Start(userB, chatB), "Can't start user")

	require.NoError(suite.T(), suite.bot.Search(userA), "Can't start user")
	require.NoError(suite.T(), suite.bot.Search(userB), "Can't start user")

	ok, err := suite.bot.JoinRequests()
	require.NoError(suite.T(), err, "Join error")
	suite.True(ok, "Requests not joined")
	messages, err := suite.ms.Next10()
	require.NoError(suite.T(), err)
	suite.Equal(4, len(messages), "Incorrect message count")
}
