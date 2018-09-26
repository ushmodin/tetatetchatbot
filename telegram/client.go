package telegram

type TelegramClient struct {
	appKey string
}

func NewTelegramClient(appKey string) (*TelegramClient, error) {
	return &TelegramClient{appKey: appKey}, nil
}

func (client TelegramClient) SendMessage() error {
	return nil
}
