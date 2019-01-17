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
	state string
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
func (group *Group) Add(line, state string) (*Status, error) {
	newStatus := &Status{
		line:  line,
		state: state,
		group: group,
	}
	group.mu.Lock()
	group.statuses = append(group.statuses, newStatus)
	maxLen := maxLength(group.statuses)
	for _, status := range group.statuses {
		status.ensureLength(maxLen)
	}
	group.mu.Unlock()
	return newStatus, group.render()
}

func (group *Group) render() error {
	group.mu.Lock()
	defer group.mu.Unlock()
	for _, status := range group.statuses {
		if _, err := fmt.Fprintln(group.writer, status); err != nil {
			return err
		}
	}
	return group.writer.Flush(len(group.statuses))
}

func (status *Status) ensureLength(maxlen int) {
	myLen := len(status.line)
	if myLen < maxlen {
		status.line = pad(status.line, maxlen-myLen, " ")
	}
}

// Set updates the status line
func (status *Status) Set(state string) error {
	status.group.mu.Lock()
	status.state = state
	status.group.mu.Unlock()
	return status.group.render()
}

func (status *Status) String() string {
	return fmt.Sprintf("%v\t[%v]", status.line, status.state)
}

func times(str string, n int) (out string) {
	return
}

// Left left-pads the string with pad up to len runes
// len may be exceeded if
func Left(str string, length int, pad string) string {
	return times(pad, length-len(str)) + str
}

// Right right-pads the string with pad up to len runes
func pad(str string, length int, pad string) string {
	for i := 0; i < length; i++ {
		str += pad
	}
	return str
}

func maxLength(statuses []*Status) int {
	maxLen := 0
	for _, status := range statuses {
		maxLen = max(maxLen, len(status.line))
	}
	return maxLen
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}
