package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/zelenin/go-tdlib/client"
)

func main() {
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	var (
		apiIdRaw = os.Getenv("API_ID")
		apiHash  = os.Getenv("API_HASH")
	)

	apiId64, err := strconv.ParseInt(apiIdRaw, 10, 32)
	if err != nil {
		log.Fatalf("strconv.Atoi error: %s", err)
	}

	apiId := int32(apiId64)

	authorizer.TdlibParameters <- &client.TdlibParameters{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(".tdlib", "database"),
		FilesDirectory:         filepath.Join(".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  apiId,
		ApiHash:                apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %s", err)
	}

	// Get a tdlibClient for communication with Telegram API
	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	// Test — getting version of the TDLib, expected: 1.8.0
	optionValue, err := tdlibClient.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

	// Test — getting information about authorized user, expected: Name Surname [username]
	me, err := tdlibClient.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Username)

	// Handle Ctrl+C , Gracefully exit and shutdown tdlib
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		tdlibClient.Stop()
		os.Exit(1)
	}()

	// Receive updates
	listener := tdlibClient.GetListener()
	defer listener.Close()

	for update := range listener.Updates {

		// We are handling only updates with new messages
		if update.GetType() != "updateNewMessage" {
			continue
		}

		// Get the content of the message and check that it's a text message
		content := update.(*client.UpdateNewMessage).Message.Content
		if content.MessageContentType() != "messageText" {
			continue
		}

		log.Printf("Message: %#v", update.(*client.UpdateNewMessage).Message.Content.(*client.MessageText).Text.Text)
		log.Printf("Text Entities: %#v", update.(*client.UpdateNewMessage).Message.Content.(*client.MessageText).Text.Entities)
	}
}
