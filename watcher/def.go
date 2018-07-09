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

type Warden struct {
	PS  map[string]*ps.PS
	Seq *goseq.MemSequencer
}

func (w *Warden) NewProcess(p *ps.PS) (int64, error) {
	go func() {
		if err := p.Start(); err != nil {
			println("Failed to start process", err.Error())
		}
	}()
	seqID, err := w.Seq.Next()
	if err != nil {
		return -1, nil
	}
	p.ID = uint(seqID)
	w.PS[strconv.FormatInt(seqID, 10)] = p
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		if p.CMD() != nil && p.CMD().Process != nil {
			break
		}
	}
	return seqID, nil
}

func (w *Warden) Processes() map[string]*ps.PS {
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

func (w *Warden) FindPSByID(id string) (*ps.PS, error) {
	for key := range w.PS {
		println(key)
	}
	if ps, exists := w.PS[id]; exists {
		return ps, nil
	}
	return nil, nil
}
