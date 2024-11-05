package main

import (
	"encoding/json"
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
	sendNotification("Welcome")
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	log.SetOutput(file)

	mainthread.Init(run)
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

func run() {
	hk := hotkey.New(getModifiers(), getKey())
	err := hk.Register()
	if err != nil {
		log.Printf("hotkey: failed to register hotkey: %v", err)
		return
	}

	for {
		<-hk.Keydown()

		response, err := http.Post("http://192.168.1.1:4782/reconnect", "", nil)
		if err != nil {
			handleError("Error executing request", err)
			continue
		}
		defer response.Body.Close()

		var respData Response
		if err := json.NewDecoder(response.Body).Decode(&respData); err != nil {
			handleError("Error parsing response", err)
			continue
		}

		if response.StatusCode == http.StatusOK {
			log.Println("Request completed successfully")
		} else {
			log.Printf("Error: status %s\n", response.Status)
		}
		message := "Successfully! Old IP: " + respData.OldIP + ", New IP: " + respData.NewIP
		sendNotification(message)
	}

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
