package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/FrontMage/goseq"
	"github.com/FrontMage/gosock"
	"github.com/FrontMage/pm/ps"
	"github.com/FrontMage/pm/server/protocol"
	"github.com/FrontMage/pm/watcher"
)

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
		seqID, err := w.NewProcess(p)
		println("New process", seqID)
		if err != nil {
			println("New process error:", err.Error())
		}

		println("Sending brief to client...")
		if err := writeBrief(conn); err != nil {
			println("Write error:", err.Error())
		}
	}
}

func writePid2File() error {
	pid := os.Getpid()
	f, err := os.Create(protocol.PidFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(strconv.Itoa(pid)))
	return err
}

func main() {
	println("- - - - - - - - - - - - - - -")
	println("server started")
	writePid2File()
	gosock.Listen("/tmp/pm.sock", switchCommand)
}
