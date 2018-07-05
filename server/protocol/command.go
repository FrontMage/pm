package protocol

const (
	CommandList  = "list"
	CommandStart = "start"
	PidFile      = "/tmp/pm.pid"
	LogFile      = "/tmp/pm.log"
	SockFile     = "/tmp/pm.sock"
)

type UpCommingCommand struct {
	Command     string   `json:"command"`
	CommandExec string   `json:"command_exec"`
	Args        []string `json:"args"`
}
