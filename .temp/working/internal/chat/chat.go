package chat

// this is sent in pubsub message body
// gets marshalled and unmarshalled
type Message struct {
	Message    string
	SenderID   string
	SenderName string
}