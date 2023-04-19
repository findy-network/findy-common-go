package fsm

type BackendData struct {
	// these two are the header part
	FromConnID string
	// ToConnID   string // todo: this is something that we don't need?
	Protocol string

	// for the start we have only string content, but maybe later..
	// see the EventData
	Subject string // this could be used for the chat room,
	// todo: but who should know it, i.e. keep track of the room
	Content string
}

type BackendChan = chan BackendData
type BackendInChan = <-chan BackendData
type BackendOutChan = chan<- BackendData