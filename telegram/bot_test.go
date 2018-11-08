package telegram

import (
	"math/rand"
	"testing"
	"time"

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

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (suite *BotSuite) SetupTest() {
	var err error
	suite.db, err = NewMgoDb("localhost", "testtetatet")
	require.NoError(suite.T(), err)
	suite.ms, err = NewBoltMessageService()
	require.NoError(suite.T(), err)
	suite.bot, err = NewBot(suite.db, suite.ms)
	require.NoError(suite.T(), err)
	rand.Seed(time.Now().UnixNano())
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
		FirstName:    randString(5),
		LastName:     randString(8),
		UserName:     randString(15),
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
		FirstName:    randString(5),
		LastName:     randString(8),
		UserName:     randString(15),
		LanguageCode: "RU",
		IsBot:        false,
	}
	chatA := Chat{
		ID: rand.Int63(),
	}
	userB := User{
		ID:           rand.Int(),
		FirstName:    randString(5),
		LastName:     randString(8),
		UserName:     randString(15),
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

func (suite *BotSuite) Test100UserSearchPauseChat() {
	count := 50
	type TestUser struct {
		user     User
		chat     Chat
		msgCount int
		pause    bool
	}
	go suite.bot.JoinRequestsLoop()

	users := make([]TestUser, count)
	for i := 0; i < count; i++ {
		users[i].user = User{
			ID:           i,
			FirstName:    randString(5),
			LastName:     randString(8),
			UserName:     randString(15),
			LanguageCode: "RU",
			IsBot:        false,
		}
		users[i].chat = Chat{
			ID: int64(i),
		}
		users[i].msgCount = 3 + rand.Intn(2)
		users[i].pause = false
		require.NoError(suite.T(), suite.bot.Start(users[i].user, users[i].chat), "Can't start user")
		require.NoError(suite.T(), suite.bot.Search(users[i].user), "Can't start user search")
	}
	time.Sleep(1 * time.Second)
	for i := 0; i < 100; i++ {
		for userIdx := 0; userIdx < count; userIdx++ {
			companyID, err := suite.bot.GetCurrentCompany(users[userIdx].user)
			require.NoError(suite.T(), err, "error while get current company")
			if users[userIdx].pause {
				require.NoError(suite.T(), suite.bot.Search(users[userIdx].user), "Search command error")
				users[userIdx].pause = false
			} else if companyID == 0 || users[userIdx].msgCount <= 0 {
				if rand.Intn(3) == 0 {
					require.NoError(suite.T(), suite.bot.Pause(users[userIdx].user), "Pause error error")
					users[userIdx].pause = true
				} else {
					require.NoError(suite.T(), suite.bot.Search(users[userIdx].user), "Search command error")
				}
			} else {
				users[userIdx].msgCount--
			}
		}
	}
}
