package protocol

const (
	CommandList = "list"
)

type UpCommingCommand struct {
	Command string `json:"command"`
}
