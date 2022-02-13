package server

type Event string

const (
	JOIN         Event = "JOIN"
	MESSAGE      Event = "MESSAGE"
	ERROR        Event = "ERROR"
	INFO         Event = "INFO"
	LEAVE        Event = "LEAVE"
	SET_USERNAME Event = "SET_USERNAME"
	STOCK        Event = "STOCK"
)

type Message struct {
	Sender  *Client `json:"sender"`
	Header  Event   `json:"header"`
	Payload string  `json:"payload"`
}

type RabbitMQMessage struct {
	MsgFor  string `json:"msg_for"`
	Payload string `json:"payload"`
}
