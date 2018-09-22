package events

import (
	"sync"
	"sync/atomic"
)

type listeners struct {
	e       *Events
	eventID string

	idCount   uint64
	lMap      map[uint64]*Listener
	lMapMutex sync.Mutex

	activeChan chan bool
}

func newListeners(closeChan <-chan struct{}, e *Events, eventID string) *listeners {
	ls := &listeners{
		e:          e,
		eventID:    eventID,
		lMap:       make(map[uint64]*Listener),
		activeChan: make(chan bool, 1),
	}

	// Start the active routine that takes care of switching
	go ls.activeRoutine(closeChan)

	return ls
}

func (ls *listeners) Add(l *Listener) {
	// Create a new listener ID and ensure it is unqiue.
	// Add it to the listeners map and set the ID.
	//
	// WARNING: Possible loop, if more than 2^64 listeners
	// have been registered. Refactor in 25 years.
	var id uint64

	for {
		if _, ok := ls.lMap[id]; !ok {
			break
		}

		id = atomic.AddUint64(&ls.idCount, 1)
	}

	l.id = id
	ls.lMap[id] = l
}

func (ls *listeners) Remove(id uint64) {
	ls.lMapMutex.Lock()
	delete(ls.lMap, id)

	// Deactivate the event if no listeners are left
	if len(ls.lMap) == 0 {
		ls.activeChan <- false
	}
	ls.lMapMutex.Unlock()
}

func (ls *listeners) activeRoutine(closeChan <-chan struct{}) {
	var (
		err error
		active bool
	)

	for {
		select {
		case <-closeChan:
			return

		case active = <-ls.activeChan:
			err = ls.e.callSetEvent(ls.eventID, active)
			if err != nil {
				return
			}
		}
	}
}
