package websocket

import coderws "github.com/coder/websocket"

type Client struct {
	UserID  string
	Role    string
	Country string
	Conn    *coderws.Conn
}