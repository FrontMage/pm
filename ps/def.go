package ps

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusErrored = "errored"
)

// Process defines process structure, aims to be easy to use
type Process interface {
	CMD() *exec.Cmd
	Start() error
	Stop() error
	Kill() error
	Wait() error

	Name() string
	UpTime() time.Duration
	Memory() int
	User() (string, error)

	Brief() ([]map[string]string, error)
}

// PS os process wrapper
type PS struct {
	ID          uint
	Cmd         *exec.Cmd
	Status      string
	StartTime   *time.Time
	StopTime    *time.Time
	Command     string
	Args        []string
	StdErr      io.Writer
	StdOut      io.Writer
	ProcessName string
	UID         int
}

func (ps *PS) CMD() *exec.Cmd {
	return ps.Cmd
}

func (ps *PS) Start() error {
	ps.Cmd = exec.Command(ps.Command, ps.Args...)

	if ps.StdOut != nil {
		ps.Cmd.Stdout = ps.StdOut
	}
	if ps.StdErr != nil {
		ps.Cmd.Stderr = ps.StdErr
	}

	ps.UID = os.Getuid()
	now := time.Now()
	ps.StartTime = &now
	ps.Status = StatusRunning
	return ps.Cmd.Run()
}

func (ps *PS) Stop() error {
	if ps.Cmd != nil && ps.Cmd.Process != nil {
		err := ps.Cmd.Process.Signal(os.Interrupt)
		if err != nil {
			ps.Status = StatusErrored
			return err
		} else {
			ps.Status = StatusStopped
		}
	} else {
		ps.Status = StatusStopped
	}
	now := time.Now()
	ps.StopTime = &now
	return nil
}

func (ps *PS) Kill() error {
	err := ps.Cmd.Process.Kill()
	if err != nil {
		ps.Status = StatusErrored
	} else {
		ps.Status = StatusStopped
	}
	now := time.Now()
	ps.StopTime = &now
	return err
}

func (ps *PS) Wait() error {
	err := ps.Cmd.Wait()
	if err == nil {
		ps.Status = StatusStopped
	} else {
		ps.Status = StatusErrored
	}
	return err
}

func (ps *PS) Name() string {
	return ps.ProcessName
}

func (ps *PS) UpTime() time.Duration {
	if ps.StopTime != nil {
		return time.Since(*ps.StopTime).Round(time.Second) - time.Since(*ps.StartTime).Round(time.Second)
	}
	return time.Since(*ps.StartTime).Round(time.Second)
}

func (ps *PS) Memory() int {
	panic("not implemented")
}

func (ps *PS) User() (string, error) {
	u, err := user.LookupId(strconv.Itoa(ps.UID))
	if err != nil {
		return "", err
	}
	return u.Name, nil
}

func (ps *PS) Brief() ([]map[string]string, error) {
	brief := []map[string]string{}
	user, err := ps.User()
	if err != nil {
		return nil, err
	}
	brief = append(
		brief,
		map[string]string{"id": fmt.Sprintf("%d", ps.ID)},
		map[string]string{"name": ps.Name()},
		map[string]string{"pid": fmt.Sprintf("%d", ps.Cmd.Process.Pid)},
		map[string]string{"command": ps.Command},
		// map[string]string{"args": strings.Join(ps.Args, " ")},
		map[string]string{"up time": fmtDuration(ps.UpTime())},
		map[string]string{"status": ps.Status},
		map[string]string{"user": user},
	)
	return brief, nil
}

func fmtDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
