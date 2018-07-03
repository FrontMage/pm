package watcher

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/FrontMage/goseq"
	"github.com/FrontMage/pm/ps"
	"github.com/olekukonko/tablewriter"
)

// Watcher watch over processes, keep trace there status
type Watcher interface {
	Processes() map[string]ps.Process
	// Brief brief on all processes
	Brief(writer io.Writer) error
	// BriefJSON returns a json formated brief, which should be an Array
	BriefJSON() ([]byte, error)
	// NewProcess start a new process, give a sequence id to it
	NewProcess(ps.Process) (int64, error)
}

type Warden struct {
	PS  map[string]ps.Process
	Seq *goseq.MemSequencer
}

func (w *Warden) NewProcess(p ps.Process) (int64, error) {
	go func() {
		if err := p.Start(); err != nil {
			println("Failed to start process", err.Error())
		}
	}()
	time.Sleep(2 * time.Second)
	seqID, err := w.Seq.Next()
	if err != nil {
		return -1, nil
	}
	w.PS[strconv.FormatInt(seqID, 10)] = p
	return seqID, nil
}

func (w *Warden) Processes() map[string]ps.Process {
	return w.PS
}

func (w *Warden) Brief(writer io.Writer) error {
	briefs := [][]string{}
	for _, p := range w.PS {
		b, err := p.Brief()
		if err != nil {
			return err
		}
		bSlice := []string{}
		for _, kv := range b {
			for _, v := range kv {
				bSlice = append(bSlice, v)
			}
		}
		briefs = append(briefs, bSlice)
	}
	keys := []string{}
	for _, p := range w.PS {
		b, err := p.Brief()
		if err != nil {
			return err
		}
		for _, kv := range b {
			for key := range kv {
				keys = append(keys, key)
			}
		}
		break
	}

	table := tablewriter.NewWriter(writer)
	table.SetHeader(keys)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(briefs)
	table.Render()
	return nil
}

func (w *Warden) BriefJSON() ([]byte, error) {
	briefs := []map[string]string{}
	for _, p := range w.PS {
		b, err := p.Brief()
		if err != nil {
			return nil, err
		}
		briefMap := map[string]string{}
		for _, kv := range b {
			for key, v := range kv {
				briefMap[key] = v
			}
			briefs = append(briefs, briefMap)
		}
	}

	return json.Marshal(briefs)
}
