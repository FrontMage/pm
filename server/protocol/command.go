package protocol

const (
	CommandList  = "list"
	CommandStart = "start"
)

type UpCommingCommand struct {
	Command     string   `json:"command"`
	CommandExec string   `json:"command_exec"`
	Args        []string `json:"args"`
}
