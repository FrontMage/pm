package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/FrontMage/pm/ps"
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
	case "start":
		// TODO: parse "tail -f" like string based command
		if len(os.Args) < 3 {
			println("Command is required fo start")
			return
		}
		if err := startProcess(os.Args[2:]); err != nil {
			println(err.Error())
		}
	case "kill":
		f, err := os.Open(protocol.PidFile)
		if err != nil {
			fmt.Println("pm_server daemon is not running")
			return
		}
		defer f.Close()
		if content, err := ioutil.ReadAll(f); err != nil {
			println("Failed to read from pid file", err.Error())
		} else {
			println("Found pm_server pid", string(content))
			pid, err := strconv.ParseInt(string(content), 10, 32)
			if err != nil {
				println("Parse pid failed", err.Error())
				return
			}
			p, err := os.FindProcess(int(pid))
			if err != nil {
				println("Find daemon failed", err.Error())
			}
			if err := p.Signal(os.Interrupt); err != nil {
				println("Kill failed", err.Error())
			} else {
				println("Daemon is closed")
				if err := rmIfExists(protocol.PidFile); err != nil {
					println("Failed to clean up pid file", err.Error())
				}
			}
		}
	default:
		flag.PrintDefaults()
	}

}

func rmIfExists(file string) error {
	if _, err := os.Stat(file); err == nil {
		return os.Remove(file)
	}
	return nil
}

func writeCommand(c net.Conn, co *protocol.UpCommingCommand) error {
	bytes, err := json.Marshal(co)
	if err != nil {
		return err
	}
	_, err = c.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func startProcess(commands []string) error {
	// Dial socket
	c, err := ensureSock()
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
		Command:     protocol.CommandStart,
		CommandExec: commands[0],
		Args:        commands[1:],
	}
	if err := writeCommand(c, command); err != nil {
		return err
	}

	// Listen to response
	return readResult(c)
}

func killPS(pid string) error {
	killCommand := &ps.PS{
		Command: "kill",
		Args:    []string{pid},
	}
	return killCommand.Start()
}

func ensureSock() (net.Conn, error) {
	// Dial socket
	d := net.Dialer{Timeout: time.Second}
	c, err := d.Dial("unix", "/tmp/pm.sock")
	if err != nil {
		// Socket error start the pm daemon
		fmt.Println("pm daemon seems not running, starting daemon...")
		daemon := &ps.PS{
			ProcessName: "pm server",
			Command:     "nohup",
			Args: []string{
				"pm_server",
				"&>",
				"/tmp/pm.log&",
			},
		}
		go func() {
			if err := daemon.Start(); err != nil {
				println(err.Error())
			}
		}()
		println("Daemon started")
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			conn, err := net.Dial("unix", "/tmp/pm.sock")
			if err == nil {
				return conn, nil
			}
		}
		println("Dial timeout after 10s")
	}
	return c, nil
}

func readResult(c net.Conn) error {
	// Listen to response
	buf := make([]byte, 1024)
	n, err := c.Read(buf[:])
	if err != nil {
		return err
	}
	println(string(buf[0:n]))
	return nil
}

func listProcesses() error {
	// Dial socket
	c, err := ensureSock()
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
	return readResult(c)
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
