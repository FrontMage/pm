package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"

	"github.com/FrontMage/gosock"
	"github.com/FrontMage/pm/ps"
	"github.com/FrontMage/pm/server/protocol"
	"github.com/FrontMage/pm/watcher"
	daemon "github.com/sevlyar/go-daemon"
)

// TODO: remove these, change into `start` command
var p = &ps.PS{
	ProcessName: "MyPS",
	Command:     "tail",
	Args: []string{
		"-f",
		"/var/log/install.log",
	},
	// StdErr: os.Stderr,
	// StdOut: os.Stdout,
}
var w = &watcher.Warden{
	PS: map[string]ps.Process{"0": p},
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

	switch command.Command {
	case protocol.CommandList:
		var buf bytes.Buffer
		err := w.Brief(&buf)
		if err != nil {
			println("Brief error:", err.Error())
		}
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			println("Write error:", err.Error())
		}
	}
}

func main() {
	// TODO: remove these
	go p.Start()
	defer p.Stop()

	cntxt := &daemon.Context{
		PidFileName: "/tmp/pm.pid",
		PidFilePerm: 0644,
		LogFileName: "/tmp/pm.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		// Args:        []string{"[go-daemon sample]"},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	gosock.Listen("/tmp/pm.sock", switchCommand)
}
