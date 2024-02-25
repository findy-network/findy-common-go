package fsm

// BackendData is important value object to transoprt data between f-fsm and
// b-fsm. If these values are added remember implement their handling and
// copying to f-fsm's Memory.
// See Event.copyBackendDataValuesToMemory(), and event building
type BackendData struct {
	// these two are the header part
	ConnID   string // same as conversation FSM connID
	Protocol string

	SessionID string // this is filtering part when available

	// for the start we have only string content, but maybe later..
	// see the EventData
	Subject string // this could be used for the chat room,

	// todo: but who should know it, i.e. keep track of the room
	Content string
}

type BackendChan = chan *BackendData
type BackendInChan = <-chan *BackendData
type BackendOutChan = chan<- *BackendData
