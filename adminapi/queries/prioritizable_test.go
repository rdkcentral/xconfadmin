package queries

import (
	"testing"

	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/stretchr/testify/assert"
)

type mockPrioritizable struct {
	id       string
	priority int
}

func (m *mockPrioritizable) GetID() string     { return m.id }
func (m *mockPrioritizable) GetPriority() int  { return m.priority }
func (m *mockPrioritizable) SetPriority(p int) { m.priority = p }

func TestUpdatePrioritizablePriorityAndReorganize(t *testing.T) {
	t.Parallel()
	item1 := &mockPrioritizable{id: "a", priority: 1}
	item2 := &mockPrioritizable{id: "b", priority: 2}
	item3 := &mockPrioritizable{id: "c", priority: 3}
	items := []core.Prioritizable{item1, item2, item3}
	newItem := &mockPrioritizable{id: "b", priority: 1}
	result := UpdatePrioritizablePriorityAndReorganize(newItem, items, 2)
	assert.NotNil(t, result)
	assert.Equal(t, newItem.priority, 1)
}

func TestPackPriorities(t *testing.T) {
	t.Parallel()
	item1 := &mockPrioritizable{id: "a", priority: 1}
	item2 := &mockPrioritizable{id: "b", priority: 2}
	item3 := &mockPrioritizable{id: "c", priority: 3}
	items := []core.Prioritizable{item1, item2, item3}
	altered := PackPriorities(items, item2)
	assert.NotNil(t, altered)
	// After deleting item2 (priority 2), only item3 changes priority from 3 to 2
	// item1 stays at priority 1 (unchanged, not in altered list)
	assert.Equal(t, 1, len(altered))
	assert.Equal(t, 2, altered[0].GetPriority())
	assert.Equal(t, "c", altered[0].GetID())
}

func TestPackPriorities_ErrorCase(t *testing.T) {
	t.Parallel()
	item1 := &mockPrioritizable{id: "a", priority: 1}
	item2 := &mockPrioritizable{id: "b", priority: 2}
	items := []core.Prioritizable{item1, item2}
	// Try to delete an item not in the list
	itemNotExist := &mockPrioritizable{id: "x", priority: 99}
	altered := PackPriorities(items, itemNotExist)
	assert.NotNil(t, altered)
	assert.Equal(t, len(altered), 0)
}
