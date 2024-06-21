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

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var broker = os.Getenv("BROKER")
	var port = os.Getenv("PORT")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
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

	// キーボード入力の初期化
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	fmt.Println("Press ESC to quit")
	fmt.Println("w: forward, s: backward, a: left, d: right, j: stop")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		var json string

		switch char {
		case 'w':
			json = `{"data": "forward"}`
			fmt.Println("w")
		case 's':
			json = `{"data": "backward"}`
			fmt.Println("s")
		case 'a':
			json = `{"data": "left"}`
			fmt.Println("a")
		case 'd':
			json = `{"data": "right"}`
			fmt.Println("d")
		case 'j':
			json = `{"data": "stop"}`
			fmt.Println("j")
		default:
			if key == keyboard.KeyEsc {
				client.Disconnect(250)
				fmt.Println("Disconnected")
				return
			}
			continue
		}

		publish(client, json)
	}
}

func publish(client mqtt.Client, data string) {
	token := client.Publish(os.Getenv("TOPIC"), 0, false, data)
	token.Wait()
	time.Sleep(time.Second)
}
