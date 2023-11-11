package main

import (
	"log"
	"net"
)

const (
	PORT             = "6969"
	DELAY_IN_SECONDS = 10
)

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	ClientDisconnected
	NewMessage
)

type Client struct {
	Conn net.Conn
	Name string
}

type Message struct {
	Text   string
	Client Client
	Type   MessageType
}

var clients map[string]Client = make(map[string]Client)

func createMessage(Text string, Client Client, Type MessageType) Message {
	return Message{
		Text:   Text,
		Type:   Type,
		Client: Client,
	}
}

func handleConnection(conn net.Conn, messages chan Message) {
	defer conn.Close()
	readBuffer := make([]byte, 512)
	var name string = ""

	conn.Write([]byte("Please enter your user name\n"))
	n, _ := conn.Read(readBuffer)
	name = string(readBuffer[0:n])

	if _, ok := clients[name]; ok {
		conn.Write([]byte("There can't be two swords in a sheath. Please try again"))
		return
	}

	log.Printf("user %s loggedin with %s", conn.RemoteAddr(), name)

	// creating a new client for the connection
	newClient := Client{
		Name: name,
		Conn: conn,
	}

	messages <- Message{
		Type:   ClientConnected,
		Client: newClient,
	}

	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			messages <- createMessage("", newClient, ClientDisconnected)
			return
		}
		// broadcating user connected and adding it to connection list
		messages <- createMessage(string(readBuffer[0:n]), newClient, NewMessage)
	}
}

func chatServer(messages chan Message) {
	for {
		msg := <-messages
		switch msg.Type {
		case ClientConnected:
			clients[msg.Client.Name] = msg.Client
		case ClientDisconnected:
			delete(clients, msg.Client.Name)
		case NewMessage:
			for k, client := range clients {
				if k != msg.Client.Name {
					message := "-----\n" + msg.Client.Name + "sent: " + msg.Text + "-----\n"
					_, err := client.Conn.Write([]byte(message))
					if err != nil {
						log.Printf("Oops!!! could not send message to %s: %s", client.Name, err)
					}
				}
			}
		}

	}
}

func main() {
	ln, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Printf("ERROR: Listening on port %s", PORT)
		return
	}
	log.Printf("INFO: running on PORT, %s", PORT)

	messages := make(chan Message)

	// starting the chat server
	go chatServer(messages)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ERROR: Could not Accept Connection: %s", err)
			continue
		}
		log.Printf("A new client connected with - %s", conn.RemoteAddr().String())

		// seperated thread to handle messages from a client during its session
		go handleConnection(conn, messages)
	}
}
