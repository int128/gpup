package photos

import (
	"testing"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

func newItem(description string) *photoslibrary.NewMediaItem {
	return &photoslibrary.NewMediaItem{Description: description}
}

func must(t *testing.T, err error, description string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s returned error: %s", description, err)
	}
}

type triggerMock struct {
	Called bool
	Arg    []*photoslibrary.NewMediaItem
}

func (m *triggerMock) Clear() {
	m.Called = false
	m.Arg = nil
}

func (m *triggerMock) Call(arg []*photoslibrary.NewMediaItem) error {
	m.Called = true
	m.Arg = arg
	return nil
}

func TestBatchCreateBuffer_abc_d(t *testing.T) {
	var m triggerMock
	b := batchCreateBuffer{
		Size:    3,
		Trigger: m.Call,
	}

	m.Clear()
	must(t, b.Add(newItem("a")), "Add(a)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("b")), "Add(b)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("c")), "Add(c)")
	if !m.Called {
		t.Errorf("Add should call Trigger but not")
	}
	if len(m.Arg) != 3 || "a" != m.Arg[0].Description || "b" != m.Arg[1].Description || "c" != m.Arg[2].Description {
		t.Errorf("Add should call with [a,b,c] but %s", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("d")), "Add(d)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Flush(), "Flush")
	if !m.Called {
		t.Errorf("Flush should call Trigger but not")
	}
	if len(m.Arg) != 1 || "d" != m.Arg[0].Description {
		t.Errorf("Flush should call with [d] but %s", m.Arg)
	}

	m.Clear()
	must(t, b.Flush(), "Flush")
	if m.Called {
		t.Errorf("Flush should not call Trigger but called with %+v", m.Arg)
	}
}

func TestBatchCreateBuffer_abc(t *testing.T) {
	var m triggerMock
	b := batchCreateBuffer{
		Size:    3,
		Trigger: m.Call,
	}

	m.Clear()
	must(t, b.Add(newItem("a")), "Add(a)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("b")), "Add(b)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("c")), "Add(c)")
	if !m.Called {
		t.Errorf("Add should call Trigger but not")
	}
	if len(m.Arg) != 3 || "a" != m.Arg[0].Description || "b" != m.Arg[1].Description || "c" != m.Arg[2].Description {
		t.Errorf("Add should call with [a,b,c] but %s", m.Arg)
	}

	m.Clear()
	must(t, b.Flush(), "Flush")
	if m.Called {
		t.Errorf("Flush should not call Trigger but called with %+v", m.Arg)
	}
}

func TestBatchCreateBuffer_ab(t *testing.T) {
	var m triggerMock
	b := batchCreateBuffer{
		Size:    3,
		Trigger: m.Call,
	}

	m.Clear()
	must(t, b.Add(newItem("a")), "Add(a)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Add(newItem("b")), "Add(b)")
	if m.Called {
		t.Errorf("Add should not call Trigger but called with %+v", m.Arg)
	}

	m.Clear()
	must(t, b.Flush(), "Flush")
	if !m.Called {
		t.Errorf("Flush should call Trigger but not")
	}
	if len(m.Arg) != 2 || "a" != m.Arg[0].Description || "b" != m.Arg[1].Description {
		t.Errorf("Flush should call with [a,b] but %s", m.Arg)
	}
}

func TestBatchCreateBuffer_empty(t *testing.T) {
	var m triggerMock
	b := batchCreateBuffer{
		Size:    3,
		Trigger: m.Call,
	}

	m.Clear()
	must(t, b.Flush(), "Flush")
	if m.Called {
		t.Errorf("Flush should not call Trigger but called with %+v", m.Arg)
	}
}
