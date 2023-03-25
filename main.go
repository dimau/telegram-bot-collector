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

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	optionValue, err := tdlibClient.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Username)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		tdlibClient.Stop()
		os.Exit(1)
	}()
}

//package main
//
//import (
//	"encoding/json"
//	"flag"
//	"fmt"
//	"github.com/Arman92/go-tdlib/v2/client"
//	"github.com/Arman92/go-tdlib/v2/tdlib"
//	"io"
//	"net/http"
//	"os"
//	"os/signal"
//	"syscall"
//	"time"
//)
//
//var clientTdLib *client.Client
//
//func main() {
//	// Get parameters from CLI
//	apiId := flag.String("API_ID", "", "API ID for our Telegram Application")
//	apiHash := flag.String("API_HASH", "", "API HASH for our Telegram Application")
//	flag.Parse()
//
//	client.SetLogVerbosityLevel(1)
//	client.SetFilePath("./errors.txt")
//
//	// Create new instance of go-tdlib clientTdLib
//	clientTdLib = client.NewClient(client.Config{
//		APIID:               *apiId,
//		APIHash:             *apiHash,
//		SystemLanguageCode:  "en",
//		DeviceModel:         "Server",
//		SystemVersion:       "1.0.0",
//		ApplicationVersion:  "1.0.0",
//		UseMessageDatabase:  true,
//		UseFileDatabase:     true,
//		UseChatInfoDatabase: true,
//		UseTestDataCenter:   false,
//		DatabaseDirectory:   "./tdlib-db",
//		FileDirectory:       "./tdlib-files",
//		IgnoreFileNames:     false,
//	})
//
//	// Handle Ctrl+C , Gracefully exit and shutdown tdlib
//	var ch = make(chan os.Signal, 2)
//	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
//	go func() {
//		<-ch
//		clientTdLib.DestroyInstance()
//		os.Exit(1)
//	}()
//
//	// Authorization in distinct goroutine
//	go func() {
//		for {
//			var currentState, _ = clientTdLib.Authorize()
//			if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPhoneNumberType {
//				fmt.Print("Enter phone: ")
//				var number string
//				fmt.Scanln(&number)
//				var _, err = clientTdLib.SendPhoneNumber(number)
//				if err != nil {
//					fmt.Printf("Error sending phone number: %v\n", err)
//				}
//			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitCodeType {
//				fmt.Print("Enter code: ")
//				var code string
//				fmt.Scanln(&code)
//				var _, err = clientTdLib.SendAuthCode(code)
//				if err != nil {
//					fmt.Printf("Error sending auth code : %v\n", err)
//				}
//			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPasswordType {
//				fmt.Print("Enter Password: ")
//				var password string
//				fmt.Scanln(&password)
//				var _, err = clientTdLib.SendAuthPassword(password)
//				if err != nil {
//					fmt.Printf("Error sending auth password: %v\n", err)
//				}
//			} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateReadyType {
//				fmt.Println("Authorization Ready! Let's rock")
//				break
//			}
//		}
//	}()
//
//	// Wait while we get Authorization Ready!
//	// Note: See authorization example for complete authorization sequence example
//	var currentState, _ = clientTdLib.Authorize()
//	for ; currentState.GetAuthorizationStateEnum() != tdlib.AuthorizationStateReadyType; currentState, _ = clientTdLib.Authorize() {
//		time.Sleep(300 * time.Millisecond)
//	}
//
//	http.HandleFunc("/getChats", getChatsHandler)
//	http.ListenAndServe(":3000", nil)
//}
//
//func getChatsHandler(w http.ResponseWriter, req *http.Request) {
//	var allChats, getChatErr = getChatList(clientTdLib, 1000)
//	if getChatErr != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte(getChatErr.Error()))
//		return
//	}
//
//	var retMap = make(map[string]interface{})
//	retMap["total"] = len(allChats)
//
//	var chatTitles []string
//	for _, chat := range allChats {
//		chatTitles = append(chatTitles, chat.Title)
//	}
//
//	retMap["chatList"] = chatTitles
//
//	var ret, marshalErr = json.Marshal(retMap)
//	if marshalErr != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte(marshalErr.Error()))
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	io.WriteString(w, string(ret))
//}
//
//// see https://stackoverflow.com/questions/37782348/how-to-use-getchats-in-tdlib
//func getChatList(client *client.Client, limit int) ([]*tdlib.Chat, error) {
//	var allChats []*tdlib.Chat
//	var chatList = tdlib.NewChatListMain()
//
//	chats, err := client.GetChats(chatList, int32(limit))
//	if err != nil {
//		return allChats, err
//	}
//
//	for len(chats.ChatIDs) != limit {
//		// get chats (ids) from tdlib
//		_, err := client.LoadChats(chatList, int32(limit-len(chats.ChatIDs)))
//		if err != nil {
//			if err.(tdlib.RequestError).Code != 404 {
//				chats, err = client.GetChats(chatList, int32(limit))
//				break
//			}
//			return allChats, err
//		}
//
//		chats, err = client.GetChats(chatList, int32(limit))
//	}
//
//	for _, chatID := range chats.ChatIDs {
//		// get chat info from tdlib
//		chat, err := client.GetChat(chatID)
//		if err == nil {
//			allChats = append(allChats, chat)
//		} else {
//			return allChats, err
//		}
//	}
//
//	return allChats, nil
//}
