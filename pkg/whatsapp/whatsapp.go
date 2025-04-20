package whatsapp

import (
	"ProjectGolang/database/postgres"
	"context"
	"fmt"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"time"
)

type Message struct {
	PhoneNumber string
	Text        string
}

type IWhatsappSender interface {
	SendMessage(ctx context.Context, phoneNumber, message string) error
	Disconnect() error
	IsConnected() bool
}

type whatsappSender struct {
	client *whatsmeow.Client
}

func New() (IWhatsappSender, error) {
	dsn := postgres.FormatDSN()

	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New("postgres", dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))

	connected := make(chan bool)
	client.AddEventHandler(func(evt interface{}) {
		if _, ok := evt.(*events.Connected); ok {
			connected <- true
		}
	})

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					fmt.Println("QR Code:", evt.Code)
				}
			}
		}()
	} else {
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	select {
	case <-connected:
		fmt.Println("WhatsApp connected")
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("connection timeout")
	}

	return &whatsappSender{
		client: client,
	}, nil
}

func (w *whatsappSender) SendMessage(ctx context.Context, phoneNumber, message string) error {
	jid := types.NewJID(phoneNumber, types.DefaultUserServer)

	waMsg := &waProto.Message{
		Conversation: proto.String(message),
	}

	_, err := w.client.SendMessage(ctx, jid, waMsg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (w *whatsappSender) Disconnect() error {
	w.client.Disconnect()
	return nil
}

func (w *whatsappSender) IsConnected() bool {
	return w.client.IsConnected()
}
