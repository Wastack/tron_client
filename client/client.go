package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/tron_client/types"
	"log"
	"net"
)

type JsonType struct{}

type Client struct {
	conn      net.Conn
	Msgs      chan types.JsonMsgI
	connected bool
}

func Connect(address string, port int) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}
	c := Client{
		conn:      conn,
		Msgs:      make(chan types.JsonMsgI, 1),
		connected: true,
	}
	return &c, nil
}

func (c *Client) Listen() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Printf("Listen: %s", err.Error())
			break
		}
		gen := &types.JsonMsg{}
		err = json.Unmarshal([]byte(msg), gen)
		if err != nil {
			log.Printf("Listen: missing type parameter")
			break
		}

		// TODO suuport connack, start game, server tick, error message
		switch gen.Type {
		case "ready":
			readyMsg := types.ReadyMsg{}
			err = json.Unmarshal([]byte(msg), readyMsg)
			if err != nil {
				log.Printf("Listen: malformed ready message")
				break
			}
			c.Msgs <- readyMsg
		case "chat":
			chatMsg := &types.ChatMsg{}
			err = json.Unmarshal([]byte(msg), chatMsg)
			if err != nil {
				log.Printf("Listen: malformed chat message")
				break
			}
			c.Msgs <- chatMsg
		case "connection":
			conAck := &types.ConnAckMsg{}
			err = json.Unmarshal([]byte(msg), conAck)
			if err != nil {
				log.Printf("Listen: malformed chat message")
				break
			}
			c.Msgs <- conAck
		case "start_game":
			// TODO
			log.Printf("Listen: Unkown message type")
		default:
			log.Printf("Listen: Unkown message type")
		}
	}
}

func (c *Client) Close() {
	log.Printf("Close network connection")
	c.connected = false
	c.conn.Close()
}

func (c *Client) SendMessage(message []byte) error {
	if !c.connected {
		return fmt.Errorf("Socket is closed")
	}
	c.conn.Write(append(message, '\n'))
	return nil
}

func (c *Client) ConnectRequest(name string, groupId string,
	privacy string) (*types.ConnRespMsg, error) {
	resp := &types.ConnRespMsg{}
	if !c.connected {
		return resp, fmt.Errorf("Socket is closed")
	}
	// send connection request
	log.Print("Send connect request to server")
	conReq, err := json.Marshal(types.ConnReqMsg{
		JsonMsg: &types.JsonMsg{Type: "connect"},
		Name:    "Wastack",
		Privacy: "private",
	})
	if err != nil {
		log.Fatal("Failed to marshal connect message")
	}
	c.SendMessage(conReq)

	log.Print("Receiving connect response")
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		log.Printf("Connection error: %s", err.Error())
	}
	err = json.Unmarshal([]byte(msg), resp)
	if err != nil {
		log.Printf("Unexpected msg from server: %s", msg)
		return resp, err
	}
	log.Print("Connect response received")
	return resp, nil
}
