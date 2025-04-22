package main

import (
	"encoding/json"
	"fmt"
	"github.com/gen2brain/beeep"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
	"log"
	"net/http"
	"os"
)

type Response struct {
	OldIP    string `json:"old_ip"`
	NewIP    string `json:"new_ip"`
	Duration int64  `json:"duration"`
}

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.SetOutput(file)

	// CLI mode
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		handleCLI(os.Args[2:])
		return
	}

	// hotkey node (default)
	sendNotification("Welcome")
	mainthread.Init(run)
}

func handleCLI(args []string) {
	if len(args) == 0 {
		fmt.Println("Available commands: reconnect, help")
		return
	}

	switch args[0] {
	case "reconnect":
		if err := doReconnect(); err != nil {
			handleError("Reconnect error", err)
		}
	case "help":
		fmt.Println("Commands:\n  reconnect - trigger reconnect\n  help - show this help")
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

func doReconnect() error {
	response, err := http.Post("http://192.168.1.1:4782/reconnect", "", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var respData Response
	if err := json.NewDecoder(response.Body).Decode(&respData); err != nil {
		return err
	}

	if response.StatusCode == http.StatusOK {
		log.Println("Reconnect completed successfully")
		message := fmt.Sprintf("Successfully! Old IP: %s, New IP: %s", respData.OldIP, respData.NewIP)
		sendNotification(message)
	} else {
		log.Printf("Error: status %s\n", response.Status)
	}
	return nil
}

func run() {
	hk := hotkey.New(getModifiers(), getKey())
	if err := hk.Register(); err != nil {
		log.Printf("hotkey: failed to register hotkey: %v", err)
		return
	}

	for {
		<-hk.Keydown()
		if err := doReconnect(); err != nil {
			handleError("Error during hotkey reconnect", err)
		}
	}
}

func getModifiers() []hotkey.Modifier {
	return []hotkey.Modifier{
		hotkey.ModCtrl,
		hotkey.ModShift,
	}
}

func getKey() hotkey.Key {
	return hotkey.KeyS
}

func handleError(message string, err error) {
	sendNotification(message)
	log.Printf("%s: %v", message, err)
}

func sendNotification(message string) {
	if notifyErr := beeep.Notify("IP Changer", message, "icon.png"); notifyErr != nil {
		log.Printf("Failed to send notification: %v", notifyErr)
	}
}
