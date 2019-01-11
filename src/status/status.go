package status

import (
	"fmt"
	"sync"

	"github.com/tanema/wrp/src/colors"
	"github.com/tanema/wrp/src/cwriter"
)

// Group holds a group of statuses
type Group struct {
	writer   *cwriter.Writer
	statuses []*Status
	mu       sync.Mutex
}

// Status holds a single line status
type Status struct {
	group *Group
	line  string
}

// New creates a new status line group
func New() *Group {
	group := &Group{
		writer:   cwriter.New(colors.ColorStdOut),
		statuses: []*Status{},
	}
	return group
}

// Add a new line
func (group *Group) Add(initialLine string) (*Status, error) {
	newStatus := &Status{
		line:  initialLine,
		group: group,
	}
	group.mu.Lock()
	group.statuses = append(group.statuses, newStatus)
	group.mu.Unlock()
	return newStatus, group.render()
}

func (group *Group) render() error {
	group.mu.Lock()
	defer group.mu.Unlock()
	for _, status := range group.statuses {
		if _, err := fmt.Fprintln(group.writer, status.line); err != nil {
			return err
		}
	}
	return group.writer.Flush(len(group.statuses))
}

// Set updates the status line
func (status *Status) Set(line string) error {
	status.group.mu.Lock()
	status.line = line
	status.group.mu.Unlock()
	return status.group.render()
}
