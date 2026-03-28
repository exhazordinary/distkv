package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Message represents a Maelstrom protocol message.
// Every message has a src (sender), dest (receiver), and a body (payload).
type Message struct {
	Src  string          `json:"src"`
	Dest string          `json:"dest"`
	Body json.RawMessage `json:"body"`
}

// MessageBody represents the inner body of a message.
// "type" tells us what kind of message it is (e.g., "echo", "init").
// "msg_id" and "in_reply_to" are for request/response matching.
type MessageBody struct {
	Type      string `json:"type"`
	MsgID     int    `json:"msg_id,omitempty"`
	InReplyTo int    `json:"in_reply_to,omitempty"`
}

// Node represents our distributed node.
type Node struct {
	id string
}

func main() {
	node := &Node{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Fprintf(os.Stderr, "error unmarshaling message: %v\n", err)
			continue
		}

		var body MessageBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			fmt.Fprintf(os.Stderr, "error unmarshaling body: %v\n", err)
			continue
		}

		switch body.Type {
		case "init":
			// Maelstrom sends this first to tell us our node ID
			var initBody struct {
				MessageBody
				NodeID string `json:"node_id"`
			}
			json.Unmarshal(msg.Body, &initBody)
			node.id = initBody.NodeID

			reply := map[string]any{
				"type":        "init_ok",
				"in_reply_to": initBody.MsgID,
			}
			node.send(msg.Src, reply)

		case "echo":
			var echoBody struct {
				MessageBody
				Echo string `json:"echo"`
			}
			json.Unmarshal(msg.Body, &echoBody)

			reply := map[string]any{
				"type": "echo_ok",
				"in_reply_to": echoBody.MsgID,
				"echo": echoBody.Echo,
			}
			node.send(msg.Src, reply)
		}
	}
}

// send marshals a reply body and writes it to stdout as a complete Message.
func (n *Node) send(dest string, body map[string]any) {
	bodyBytes, _ := json.Marshal(body)
	msg := Message{
		Src:  n.id,
		Dest: dest,
		Body: bodyBytes,
	}
	out, _ := json.Marshal(msg)
	fmt.Println(string(out))
}
