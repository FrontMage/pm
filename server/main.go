package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	PS:  map[string]*ps.PS{},
	Seq: &goseq.MemSequencer{},
}

func writeBrief(conn net.Conn) error {
	var buf bytes.Buffer
	err := w.Brief(&buf)
	if err != nil {
		return err
	}
	return writeResult(conn, buf.Bytes())
}

func writeResult(conn net.Conn, result []byte) error {
	if len(result) > 0 {
		_, err := conn.Write(result)
		if err != nil {
			return err
		}
		println(string(result))
	} else {
		_, err := conn.Write([]byte("No running process"))
		if err != nil {
			return err
		}
		println(string(result))
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
			ProcessName: command.CommandName,
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
			writeResult(conn, []byte(fmt.Sprintf("New process error: %s", err.Error())))
			return
		}

		println("Sending brief to client...")
		if err := writeBrief(conn); err != nil {
			println("Write error:", err.Error())
		}
	case protocol.CommandStop:
		if ps, err := w.FindPSByID(command.CommandID); err != nil {
			println("Stop process error:", err.Error())
		} else if ps == nil {
			println("Can't find process id=", command.CommandID)
			writeResult(conn, []byte("Can't find process"))
			return
		} else if err := ps.Stop(); err != nil {
			println("Failed to stop process:", err.Error())
			writeResult(conn, []byte(fmt.Sprintf("Failed to stop process: %s", err.Error())))
			return
		} else {
			println("Sending brief to client...")
			if err := writeBrief(conn); err != nil {
				println("Write error:", err.Error())
			}
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
	gosock.Listen(protocol.SockFile, switchCommand)
}
