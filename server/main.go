package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/FrontMage/goseq"
	"github.com/FrontMage/gosock"
	"github.com/FrontMage/pm/ps"
	"github.com/FrontMage/pm/server/protocol"
	"github.com/FrontMage/pm/watcher"
	// daemon "github.com/takama/daemon"
)

var pidFile = "/tmp/pm.pid"
var pmLogFile = "/tmp/pm.log"

var w = &watcher.Warden{
	PS:  map[string]ps.Process{},
	Seq: &goseq.MemSequencer{},
}

func writeBrief(conn net.Conn) error {
	var buf bytes.Buffer
	err := w.Brief(&buf)
	if err != nil {
		return err
	}
	if buf.Len() > 0 {
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			return err
		}
		println(buf.String())
	} else {
		_, err = conn.Write([]byte("No running process"))
		if err != nil {
			return err
		}
		println(buf.String())
	}
	return nil
}

func switchCommand(conn net.Conn) {
	buf := make([]byte, 512)
	numRead, err := conn.Read(buf)
	if err != nil {
		log.Fatal("Read error:", err)
	}

	data := buf[0:numRead]

	command := &protocol.UpCommingCommand{}
	if err := json.Unmarshal(data, command); err != nil {
		println("Unmarshal command error:", err.Error())
		return
	}

	println("Receiving command:", command.Command)

	switch command.Command {
	case protocol.CommandList:
		if err := writeBrief(conn); err != nil {
			println("Write error:", err.Error())
		}
	case protocol.CommandStart:
		p := &ps.PS{
			ProcessName: "MyPS",
			Command:     command.CommandExec,
			Args:        command.Args,
			// StdErr: os.Stderr,
			// StdOut: os.Stdout,
		}
		println("Starting process...")
		go func() {
			_, err := w.NewProcess(p)
			if err != nil {
				println("New process error:", err.Error())
			}
		}()

		time.Sleep(2 * time.Second)
		println("Sending brief to client...")
		if err := writeBrief(conn); err != nil {
			println("Write error:", err.Error())
		}
	}
}

func rmIfExists(file string) error {
	if _, err := os.Stat(file); err == nil {
		return os.Remove(file)
	}
	return nil
}

func main() {
	// service, err := daemon.New("pm_server", "go process manager server daemon")
	// if err != nil {
	// 	log.Fatal("Error: ", err)
	// }
	// status, err := switchArgs(service)
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }

	// fmt.Println(status)
	// go gosock.Listen("/tmp/pm.sock", switchCommand)
	println("- - - - - - - - - - - - - - -")
	println("server started")
	gosock.Listen("/tmp/pm.sock", switchCommand)
}

// func switchArgs(service daemon.Daemon) (string, error) {
// 	switch os.Args[1] {
// 	case "install":
// 		return service.Install()
// 	case "remove":
// 		return service.Remove()
// 	case "status":
// 		return service.Status()
// 	case "start":
// 		return service.Start()
// 	case "stop":
// 		return service.Stop()
// 	}
// 	return "", nil
// }
