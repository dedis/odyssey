package helpers

import (
	"sync"

	"go.dedis.ch/onet/v3/log"
)

// StatusNotifier notify subscribers about the update of a project or request
// status
type StatusNotifier struct {
	sync.Mutex
	Status string
	// This bool indicates if the StatusNotifier is done or not. Once set to
	// true it shoudl never be set back to false
	Terminated  bool
	Subscribers []*StatusNotifierSubscriber
	ID          string
}

// StatusNotifierSubscriber represent a client that subscribes to receive
// notifications about a status
type StatusNotifierSubscriber struct {
	NotifyStream chan string
	Done         chan bool
}

// NewStatusNotifier return a new status notifier
func NewStatusNotifier() *StatusNotifier {
	return &StatusNotifier{
		Terminated:  false,
		Subscribers: make([]*StatusNotifierSubscriber, 0),
		ID:          RandString(16),
	}
}

// Close terminates a subscriber
func (s *StatusNotifierSubscriber) Close() {
	close(s.Done)
}

// Subscribe creates a new StatusNotifierSubscriber that will be notified of
// status update.
func (s *StatusNotifier) Subscribe() *StatusNotifierSubscriber {
	s.Lock()
	newClient := &StatusNotifierSubscriber{
		NotifyStream: make(chan string, 100),
		Done:         make(chan bool),
	}
	if s.Terminated {
		log.Info("status notifier is terminated, so we close the client")
		newClient.Close()
	} else {
		log.Info("new status notifier client registered")
		s.Subscribers = append(s.Subscribers, newClient)
	}
	s.Unlock()
	return newClient
}

// UpdateStatus updates the status and notifies all the subscribers
func (s *StatusNotifier) UpdateStatus(status string) {
	s.Lock()
	s.Status = status
	for _, client := range s.Subscribers {
		client.NotifyStream <- status
	}
	s.Unlock()
}

// UpdateStatusAndClose updates the status and notifies all the subscribers
func (s *StatusNotifier) UpdateStatusAndClose(status string) {
	s.Lock()
	s.Status = status
	for _, client := range s.Subscribers {
		client.NotifyStream <- status
	}
	s.Terminated = true
	for _, client := range s.Subscribers {
		client.Close()
	}
	s.Subscribers = make([]*StatusNotifierSubscriber, 0)
	s.Unlock()
}
