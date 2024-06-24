package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eiannone/keyboard"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	client := connectMQTT()
	defer client.Disconnect(250)

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	handleKeyboardInput(client)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func connectMQTT() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", os.Getenv("BROKER"), os.Getenv("PORT")))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(os.Getenv("TOKEN"))
	opts.SetPassword(os.Getenv("PASS"))
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

func handleKeyboardInput(client mqtt.Client) {
	fmt.Println("Press ESC to quit")
	fmt.Println("w: forward, s: backward, a: left, d: right, j: stop")

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		if key == keyboard.KeyEsc {
			fmt.Println("Disconnected")
			return
		}

		json := getCommand(char)
		if json != "" {
			publish(client, json)
		}
	}
}

func getCommand(char rune) string {
	switch char {
	case 'w':
		fmt.Println("w")
		return `{"data": "forward"}`
	case 's':
		fmt.Println("s")
		return `{"data": "backward"}`
	case 'a':
		fmt.Println("a")
		return `{"data": "left"}`
	case 'd':
		fmt.Println("d")
		return `{"data": "right"}`
	case 'j':
		fmt.Println("j")
		return `{"data": "stop"}`
	default:
		return ""
	}
}

func publish(client mqtt.Client, data string) {
	token := client.Publish(os.Getenv("TOPIC"), 0, false, data)
	token.Wait()
	time.Sleep(time.Second)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}
