package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"go.dedis.ch/onet/v3/log"
)

// TaskList are all the chans concerning tasks
// var TaskList = make([]*Task, 0)

// TaskListLock is a mutex to be used when updating the TaskList
// var TaskListLock = &sync.Mutex{}

// StatusTask ...
type StatusTask string

// TaskEventType ...
type TaskEventType string

const (
	// StatusWorking says the task is working
	StatusWorking = "working"
	// StatusFinished says the task is done
	StatusFinished = "finished"
	// StatusErrored says the task errored and is done
	StatusErrored = "errored"
	// TypeInfo is the standard event type
	TypeInfo = "info"
	// TypeError is an event type that indicates an error
	TypeError = "error"
	// TypeImportantInfo is an event type that indicates an important event
	TypeImportantInfo = "importantInfo"
	// TypeCloseOK indicates that the stream should be closed. That should be
	// used once the task ended well.
	TypeCloseOK = "closeOK"
	// TypeCloseError should be used to indicate that the stream is closed with
	// an error
	TypeCloseError = "closeError"
)

// TaskEvent is an event that happened on a specific task, stored in a chan
type TaskEvent struct {
	Type    TaskEventType `json:"type"`
	Time    string        `json:"time"`
	Message string        `json:"message"`
	Details string        `json:"details"`
	Source  string        `json:"source"`
}

// TaskEventSorter is used to define a sorter that sorts projects by the CreatedAt
// field.
type TaskEventSorter []*TaskEvent

// Len returns the len
func (p TaskEventSorter) Len() int { return len(p) }

// Swap swaps
func (p TaskEventSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less compares two projects based on their CreatedAt fields
func (p TaskEventSorter) Less(i, j int) bool {
	t1, err1 := time.Parse("02-01-2006 15:04:05..999", p[i].Time)
	t2, err2 := time.Parse("02-01-2006 15:04:05..999", p[j].Time)
	// in case we cannot parse any of the two times, we compare based on the
	// string.
	if err1 != nil || err2 != nil {
		return strings.Compare(p[i].Time, p[j].Time) < 0
	}
	return t1.Before(t2)
}

// TaskManagerI defines all the primitives needed to handle tasks.
type TaskManagerI interface {
	NewTask(title string) TaskI
	NumTasks() int
	GetTask(index int) TaskI
	GetSortedTasks() []TaskI
	DeleteAllTasks()
	RestoreTasks(tasks []TaskI) error
}

// DefaultTaskManager provides a default implementation of the TaskManager
// interface.
type DefaultTaskManager struct {
	sync.Mutex
	taskList []*Task
}

// NewDefaultTaskManager return a new DefaultTaskManager
func NewDefaultTaskManager() *DefaultTaskManager {
	return &DefaultTaskManager{
		taskList: make([]*Task, 0),
	}
}

// NewTask return a new Task
func (dtm *DefaultTaskManager) NewTask(title string) TaskI {
	dtm.Lock()
	index := len(dtm.taskList)
	idStr := RandString(16)
	task := &Task{
		ID:          idStr,
		Index:       index,
		Subscribers: make([]*Subscriber, 0),
		History:     make([]*TaskEvent, 0),
		Status:      StatusWorking,
		Description: title,
		StartD:      time.Now().Format("02-01-2006 15:04:05..999"),
		EndD:        "?",
	}
	dtm.taskList = append(dtm.taskList, task)
	dtm.Unlock()
	return task
}

// NumTasks ...
func (dtm *DefaultTaskManager) NumTasks() int {
	return len(dtm.taskList)
}

// GetTask ...
func (dtm *DefaultTaskManager) GetTask(index int) TaskI {
	return dtm.taskList[index]
}

// GetSortedTasks ...
func (dtm *DefaultTaskManager) GetSortedTasks() []TaskI {
	sorted := make([]*Task, len(dtm.taskList))
	copy(sorted, dtm.taskList)
	sort.Sort(sort.Reverse(TaskSorter(sorted)))

	result := make([]TaskI, dtm.NumTasks())
	for i, el := range sorted {
		result[i] = el
	}
	return result
}

// DeleteAllTasks deletes all the tasks
func (dtm *DefaultTaskManager) DeleteAllTasks() {
	dtm.taskList = make([]*Task, 0)
}

// RestoreTasks sets the provided tasks in the task list. The previous tasks are
// deleted.
func (dtm *DefaultTaskManager) RestoreTasks(tasks []TaskI) error {
	dtm.DeleteAllTasks()
	dtm.taskList = make([]*Task, len(tasks))

	for i, task := range tasks {
		if task.GetIndex() >= len(tasks) || task.GetIndex() < 0 {
			return fmt.Errorf("found anormal index on task: 0 > "+
				"(index) %d >= len(TaskList) %d", task.GetIndex(), len(tasks))
		}

		dtm.taskList[i] = &Task{
			Index:       task.GetIndex(),
			ID:          task.GetID(),
			Status:      task.GetStatus(),
			GoBackLink:  task.GetGobackLink(),
			Description: task.GetDescription(),
			History:     task.GetHistory(),
		}
	}

	return nil
}

// TaskI defines the primitives needed for a standard task. This abstraction
// allows us to better test our program.
type TaskI interface {
	CloseError(source, msg, details string)
	AddInfo(source, msg, details string)
	AddInfof(source, msg, details string, args ...interface{})
	CloseOK(source, msg, details string)
	AddTaskEvent(event TaskEvent)
	Subscribe() *Subscriber
	SetGobackLink(link string)
	GetIndex() int
	GetID() string
	GetStatus() StatusTask
	SetStatus(StatusTask)
	GetGobackLink() string
	GetDescription() string
	GetHistory() []*TaskEvent
	SetSubscribers([]*Subscriber)
	Marshall() ([]byte, error)
}

// Task implements a publisher/subscriber pattern that allows us to
// asynchronously handle tasks in our app.
type Task struct {
	sync.Mutex
	// A random string
	ID string
	// its index on the TaskList
	Index       int
	Subscribers []*Subscriber
	History     []*TaskEvent
	Status      StatusTask
	Description string
	StartD      string
	EndD        string
	// Should be used to allow the user going back to its task
	GoBackLink string
}

// StatusImg return an HTML image of the status
func (t *Task) StatusImg() string {
	return StatusImage(t.Status)
}

// PrepareAfterUnmarshal should be called after unmarshalling a task
func (t *Task) PrepareAfterUnmarshal() {
	t.Subscribers = make([]*Subscriber, 0)
}

// Marshall ...
func (t *Task) Marshall() ([]byte, error) {
	t.PrepareBeforeMarshal()
	taskBuf, err := json.Marshal(t)
	if err != nil {
		return nil, errors.New("failed to marshal task: " + err.Error())
	}

	return taskBuf, nil
}

// PrepareBeforeMarshal should be called before marshalling a task
func (t *Task) PrepareBeforeMarshal() {
	t.Subscribers = nil
}

// TaskSorter is used to define a sorter that sorts tasks by their Index field
type TaskSorter []*Task

// Len returns the len
func (p TaskSorter) Len() int { return len(p) }

// Swap swaps
func (p TaskSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less compares two Task based on their CreatedAt fields
func (p TaskSorter) Less(i, j int) bool {
	return p[i].Index < p[j].Index
}

// NewTask creates a task and returns it
// func NewTask(description string) *Task {
// 	TaskListLock.Lock()
// 	index := len(TaskList)
// 	idStr := RandString(16)
// 	task := &Task{
// 		ID:          idStr,
// 		Index:       index,
// 		Subscribers: make([]*Subscriber, 0),
// 		History:     make([]*TaskEvent, 0),
// 		Status:      StatusWorking,
// 		Description: description,
// 		StartD:      time.Now().Format("02-01-2006 15:04:05..999"),
// 		EndD:        "?",
// 	}
// 	TaskList = append(TaskList, task)
// 	TaskListLock.Unlock()
// 	return task
// }

// Subscriber is a client that subscribes to be notified. We store the past
// events so that the client can catch up.
type Subscriber struct {
	TaskStream chan *TaskEvent
	PastEvents []*TaskEvent
	Done       chan bool
}

// Close terminates a subscriber
func (s *Subscriber) Close() {
	close(s.Done)
}

// Subscribe creates a new client and will be updated by the task when a new
// event occurs.
func (t *Task) Subscribe() *Subscriber {
	t.Lock()
	newClient := &Subscriber{
		TaskStream: make(chan *TaskEvent, 100),
		PastEvents: t.History[0:len(t.History)],
		Done:       make(chan bool),
	}
	if t.Status != StatusWorking {
		log.Info("task is not in working state, so we close the client")
		newClient.Close()
	} else {
		// We must not add it if is already closed because we might close it
		// again with the CloseError or CloseOK functions.
		log.Info("new client registered")
		t.Subscribers = append(t.Subscribers, newClient)
	}
	t.Unlock()
	return newClient
}

// SetSubscribers sets the subscribers. It is usefull in the "before/after
// marshal"
func (t *Task) SetSubscribers(subscribers []*Subscriber) {
	t.Subscribers = subscribers
}

// AddTaskEvent adds a new event and notify the subscribers. If the event is of type
// closeOK or closeError, then the task is ended and the subscibers are closed.
func (t *Task) AddTaskEvent(event TaskEvent) {
	t.Lock()
	t.History = prependTaskEvent(t.History, &event)
	for _, client := range t.Subscribers {
		client.TaskStream <- &event
	}
	if event.Type == TypeCloseOK || event.Type == TypeCloseError {
		for _, client := range t.Subscribers {
			client.Close()
		}
		t.Subscribers = make([]*Subscriber, 0)
		t.EndD = time.Now().Format("02-01-2006 15:04:05..999")
		switch event.Type {
		case TypeCloseOK:
			t.Status = StatusFinished
		case TypeCloseError:
			t.Status = StatusErrored
		}
	}
	t.Unlock()
}

// SetGobackLink sets the GobackLink attribute
func (t *Task) SetGobackLink(link string) {
	t.GoBackLink = link
}

// GetIndex returns the index of the task
func (t *Task) GetIndex() int {
	return t.Index
}

// GetID returns the ID of the task
func (t *Task) GetID() string {
	return t.ID
}

// GetStatus returns the status of the task
func (t *Task) GetStatus() StatusTask {
	return t.Status
}

// SetStatus sets the status of the task
func (t *Task) SetStatus(status StatusTask) {
	t.Status = status
}

// GetGobackLink returns the goback link of the task
func (t *Task) GetGobackLink() string {
	return t.GoBackLink
}

// GetDescription returns the description of the task
func (t *Task) GetDescription() string {
	return t.Description
}

// GetHistory returns the history of the task
func (t *Task) GetHistory() []*TaskEvent {
	return t.History
}

// AddInfo adds a new task event of type info
func (t *Task) AddInfo(source, msg, details string) {
	t.AddTaskEvent(NewTaskEventInfo(source, msg, details))
}

// AddInfof adds a new task event of type info
func (t *Task) AddInfof(source, msg, details string, args ...interface{}) {
	t.AddTaskEvent(NewTaskEventInfof(source, msg, details, args...))
}

// CloseError sets the error status and closes all the clients
func (t *Task) CloseError(source, msg, details string) {
	t.AddTaskEvent(NewTaskEventCloseError(source, msg, details))
}

// CloseOK sets the finished status and closes all the clients
func (t *Task) CloseOK(source, msg, details string) {
	t.AddTaskEvent(NewTaskEventCloseOK(source, msg, details))
}

// JSONStream build a Json string representation following the SSE stream
// format, which starts each message with "data:" and must end with a double new
// line.
func (t TaskEvent) JSONStream() string {
	out := new(strings.Builder)
	out.WriteString("data: ")

	taskEventBuf, err := json.Marshal(t)
	if err != nil {
		out.WriteString("ERROR: FAILED TO MARSHAL THE TASKEVENT")
		// We should normally send two consecutive \n in order to say that the
		// stream of data can be sent. However since we send only a json
		// formatted task event, we don't need this feature and send the data at
		// the first \n.
		out.WriteString("\n")
		return out.String()
	}
	out.Write(taskEventBuf)
	out.WriteString("\n\n")
	return out.String()
}

// NewTaskEvent creates a new task event
func NewTaskEvent(ttype TaskEventType, source, msg, details string) TaskEvent {
	return TaskEvent{
		Type:    ttype,
		Time:    time.Now().Format("02-01-06-03:04:05..999"),
		Message: msg,
		Details: details,
		Source:  source,
	}
}

// NewTaskEventf creates a new task event with a formatted string as the details
func NewTaskEventf(ttype TaskEventType, source, msg, details string, args ...interface{}) TaskEvent {
	return TaskEvent{
		Type:    ttype,
		Time:    time.Now().Format("02-01-06-03:04:05..999"),
		Message: msg,
		Details: fmt.Sprintf(details, args...),
		Source:  source,
	}
}

// FACTORY

// TaskEventFactory is a handy struct that make the creation of task events a
// bit less tidious.
type TaskEventFactory struct {
	Source string
}

// NewTaskEventFactory creates a new TaskFactory
func NewTaskEventFactory(source string) *TaskEventFactory {
	return &TaskEventFactory{source}
}

// NewTaskEventInfo return a task to print info
func (t TaskEventFactory) NewTaskEventInfo(msg, details string) TaskEvent {
	return NewTaskEventInfo(t.Source, msg, details)
}

// NewTaskEventError return a task to print an error
func (t TaskEventFactory) NewTaskEventError(msg, details string) TaskEvent {
	return NewTaskEventError(t.Source, msg, details)
}

// NewTaskEventCloseOK return a task to print an error
func (t TaskEventFactory) NewTaskEventCloseOK(msg, details string) TaskEvent {
	return NewTaskEventCloseOK(t.Source, msg, details)
}

// NewTaskEventCloseError return a task close error
func (t TaskEventFactory) NewTaskEventCloseError(msg, details string) TaskEvent {
	return NewTaskEventCloseError(t.Source, msg, details)
}

// NewTaskEventErrorf return a task to print a formatted error message
func (t TaskEventFactory) NewTaskEventErrorf(msg, details string, args ...interface{}) TaskEvent {
	return NewTaskEventErrorf(t.Source, msg, details, args...)
}

// NewTaskEventInfof return a task to print a formatted info message
func (t TaskEventFactory) NewTaskEventInfof(msg, details string, args ...interface{}) TaskEvent {
	return NewTaskEventInfof(t.Source, msg, details, args...)
}

// RAW FUNCTIONS

// NewTaskEventInfo return a task to print info
func NewTaskEventInfo(source, msg, details string) TaskEvent {
	return NewTaskEvent(TypeInfo, source, msg, details)
}

// NewTaskEventError return a task to print an error
func NewTaskEventError(source, msg, details string) TaskEvent {
	return NewTaskEvent(TypeError, source, msg, details)
}

// NewTaskEventCloseOK return a task to print an error message
func NewTaskEventCloseOK(source, msg, details string) TaskEvent {
	return NewTaskEvent(TypeCloseOK, source, msg, details)
}

// NewTaskEventCloseError return a task to print an error message
func NewTaskEventCloseError(source, msg, details string) TaskEvent {
	return NewTaskEvent(TypeCloseError, source, msg, details)
}

// NewTaskEventCloseErrorf return a task to print an error message
func NewTaskEventCloseErrorf(source, msg, details string, args ...interface{}) TaskEvent {
	return NewTaskEventf(TypeCloseError, source, msg, details, args...)
}

// NewTaskEventImportantInfo return a task to print info
func NewTaskEventImportantInfo(source, msg, details string) TaskEvent {
	return NewTaskEvent(TypeImportantInfo, source, msg, details)
}

// NewTaskEventErrorf return a task to print a formatted error message
func NewTaskEventErrorf(source, msg, details string, args ...interface{}) TaskEvent {
	return NewTaskEventf(TypeCloseError, source, msg, details, args...)
}

// NewTaskEventInfof return a task to print a formatted info message
func NewTaskEventInfof(source, msg, details string, args ...interface{}) TaskEvent {
	return NewTaskEventf(TypeInfo, source, msg, details, args...)
}

func prependTaskEvent(x []*TaskEvent, y *TaskEvent) []*TaskEvent {
	x = append(x, nil)
	copy(x[1:], x)
	x[0] = y
	return x
}

// StatusImage return an html image corresponding to the status
func StatusImage(status StatusTask) string {
	switch status {
	case StatusWorking:
		return "<img src=\"/assets/images/status/working.gif\"/> working"
	case StatusErrored:
		return "<img src=\"/assets/images/status/errored.gif\"/> errored"
	case StatusFinished:
		return "<img src=\"/assets/images/status/finished.gif\"/> finished"
	default:
		return ""
	}
}

// FLUSH FACTORY

// TaskEventFFactory is a handy struct that make the creation of task events a
// bit less tidious with the HTPP flusher
type TaskEventFFactory struct {
	Source  string
	flusher http.Flusher
	w       http.ResponseWriter
}

// NewTaskEventFFactory creates a new TaskFactory
func NewTaskEventFFactory(source string, flusher http.Flusher,
	w http.ResponseWriter) *TaskEventFFactory {

	return &TaskEventFFactory{source, flusher, w}
}

// FlushTaskEventError creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventError(msg, details string) {
	log.Error(msg)
	fmt.Fprint(t.w, NewTaskEventError(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventErrorf creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventErrorf(msg, details string, args ...interface{}) {

	log.Errorf(msg+" - "+details, args...)
	fmt.Fprint(t.w, NewTaskEventErrorf(t.Source, msg, details, args...).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventInfo creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventInfo(msg, details string) {
	fmt.Fprint(t.w, NewTaskEventInfo(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventInfof creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventInfof(msg, details string, args ...interface{}) {
	fmt.Fprint(t.w, NewTaskEventInfof(t.Source, msg, details, args...).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventImportantInfo creates a new task of type important info and writes
// it to the responsewriter.
func (t TaskEventFFactory) FlushTaskEventImportantInfo(msg, details string) {
	fmt.Fprint(t.w, NewTaskEventImportantInfo(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventCloseOK creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventCloseOK(msg, details string) {
	fmt.Fprint(t.w, NewTaskEventCloseOK(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventCloseError creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventCloseError(msg, details string) {
	log.Error(msg)
	fmt.Fprint(t.w, NewTaskEventCloseError(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// FlushTaskEventCloseErrorf creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) FlushTaskEventCloseErrorf(msg, details string, args ...interface{}) {

	log.Errorf(msg+" - "+details, args...)
	fmt.Fprint(t.w, NewTaskEventCloseErrorf(t.Source, msg, details, args...).JSONStream())
	t.flusher.Flush()
}

// This becomes sorcellery....

// XFlushTaskEventCloseError takes an additional parameter, an error, that is
// added to the message if not nil. We need this when for example there are some
// extra steps that must be done to handle the error, like updating the project
// instance status, and those can return an error. In that case we still want to
// flush exit but if one of those extra steps failed, the user should know about
// it.
func (t TaskEventFFactory) XFlushTaskEventCloseError(err2 error, msg, details string) {
	if err2 != nil {
		log.Error("error handling step failed: ", err2)
		details = fmt.Sprintf("%s<br><span style='color:orangered'>(!) "+
			"Something went wrong while handling the error: %s</span>",
			details, err2.Error())
	}
	log.Error(msg)
	fmt.Fprint(t.w, NewTaskEventCloseError(t.Source, msg, details).JSONStream())
	t.flusher.Flush()
}

// XFlushTaskEventCloseErrorf creates a new task and writes it to the responsewriter
func (t TaskEventFFactory) XFlushTaskEventCloseErrorf(err2 error, msg, details string, args ...interface{}) {
	if err2 != nil {
		log.Error("error handling step failed: ", err2)
		details = fmt.Sprintf("%s<br><span style='color:orangered'>(!) "+
			"Something went wrong while handling the error: %s</span>",
			details, err2.Error())
	}
	log.Errorf(msg+" - "+details, args...)
	fmt.Fprint(t.w, NewTaskEventCloseErrorf(t.Source, msg, details, args...).JSONStream())
	t.flusher.Flush()
}
