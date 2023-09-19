package main

import (
	"context"
	"os"
	"os/signal"
	"os/exec"
	"syscall"
	"errors"
	"strings"

	"github.com/disgoorg/log"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var (
	token   = os.Getenv("disgo_token")
	msgContent string

	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "send",
			Description: "send SMS to a number",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "number",
					Description: "destination number",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "message",
					Description: "text to send",
					Required:    true,
				},
			},
		},
	}
)

func main() {
	log.SetLevel(log.LevelInfo)
	log.Info("starting gammu-disc")
	log.Info("disgo version: ", disgo.Version)

	client, err := disgo.New(token,
		bot.WithDefaultGateway(),
		bot.WithEventListenerFunc(commandListener),
	)
	if err != nil {
		log.Fatal("error while building disgo instance: ", err)
		return
	}

	defer client.Close(context.TODO())

	if _, err = client.Rest().SetGlobalCommands(client.ApplicationID(), commands); err != nil {
		log.Fatal("error while registering commands: ", err)
	}

	if err = client.OpenGateway(context.TODO()); err != nil {
		log.Fatal("error while connecting to gateway: ", err)
	}

	log.Infof("gammu-disc is now running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}

func commandListener(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if data.CommandName() == "send" {
		gammuErr :=  sendSMS(data.String("message"), data.String("number"))
		if (gammuErr != nil) {
			msgContent = gammuErr.Error()
		} else {
			msgContent = "SMS sent successfully"
		}
		err := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(msgContent).
			Build(),
		)
		if err != nil {
			event.Client().Logger().Error("error sending response: ", err)
		}
	}
}

func sendSMS(msg string, number string) error {
	cmd := exec.Command("timeout", "10", "gammu-smsd-inject", "TEXT", number, "-text", msg)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.Contains(string(out), "Written message") {
		log.Debug("sms sent!")
		return nil
	} else {
		return errors.New("error sending sms")
	}
}