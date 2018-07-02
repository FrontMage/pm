package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/FrontMage/pm/server/protocol"
)

func main() {
	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < 2 {
		fmt.Println("list list all proccesses")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch os.Args[1] {
	case "list":
		if err := listProcesses(); err != nil {
			println(err.Error())
		}
	default:
		flag.PrintDefaults()
	}

}

func listProcesses() error {
	// Dial socket
	c, err := net.Dial("unix", "/tmp/pm.sock")
	// TODO: no sock or timeout, start the sever daemon
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Close(); err != nil {
			println(err.Error())
		}
	}()

	// Write command
	command := &protocol.UpCommingCommand{
		Command: protocol.CommandList,
	}
	bytes, err := json.Marshal(command)
	if err != nil {
		return err
	}
	_, err = c.Write(bytes)
	if err != nil {
		return err
	}

	// Listen to response
	buf := make([]byte, 1024)
	n, err := c.Read(buf[:])
	if err != nil {
		return err
	}
	println(string(buf[0:n]))
	return nil
}

func reader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
	}
}
