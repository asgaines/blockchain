package dmaps

import "sync"

type Dmap interface {
	ReadAll() []string
	WriteMany(ss []string)
	WriteOne(s string)
	RemoveOne(s string)
	RemoveMany(ss []string)
	Len() int
}

func New() Dmap {
	dm := dmap{
		data:    make(map[string]bool),
		RWMutex: &sync.RWMutex{},
	}

	return &dm
}

type dmap struct {
	data map[string]bool
	*sync.RWMutex
}

func (dm *dmap) ReadAll() []string {
	ss := make([]string, 0, len(dm.data))
	dm.RLock()
	for s := range dm.data {
		ss = append(ss, s)
	}
	dm.RUnlock()

	return ss
}

func (dm *dmap) WriteMany(ss []string) {
	dm.Lock()
	for _, s := range ss {
		dm.data[s] = true
	}
	dm.Unlock()
}

func (dm *dmap) WriteOne(s string) {
	dm.Lock()
	dm.data[s] = true
	dm.Unlock()
}

func (dm *dmap) RemoveOne(s string) {
	dm.Lock()
	delete(dm.data, s)
	dm.Unlock()
}

func (dm *dmap) RemoveMany(ss []string) {
	dm.Lock()
	for _, s := range ss {
		delete(dm.data, s)
	}
	dm.Unlock()
}

func (dm *dmap) Len() int {
	return len(dm.data)
}
