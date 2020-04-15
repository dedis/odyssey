package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/minio/minio-go/v6"
	"go.dedis.ch/onet/v3/log"
	"golang.org/x/xerrors"
)

// CloudNotifier notify subscribers about the update of a project or request
// status
type CloudNotifier struct {
	sync.Mutex
	Status CloudNotifierStatus
	// This bool indicates if the CloudNotifier is done or not. Once set to true
	// it shoudl never be set back to false
	Terminated bool
	ID         string
	// This variable is updated with new task events.
	TaskEventCh chan TaskEvent
	// AliasName is the name used when configuring minio, for example "dedis"
	AliasName string
	// BucketName is the name of the bucket. In our case it's the project id
	BucketName string
	// This is the path after the bucket name, in the form
	// "logs/{request index}"
	Prefix      string
	IsClosed    bool
	Done        chan bool
	NotifDoneCh chan struct{}
}

// CloudNotifierSubscriber represent a client that subscribes to receive
// notifications about a cloud endpoint
type CloudNotifierSubscriber struct {
	NotifyStream chan string
	Done         chan bool
}

// CloudNotifierStatus indicates the status of the cloud notifier
type CloudNotifierStatus string

const (
	// CloudNotifierStatusCreated ...
	CloudNotifierStatusCreated = ""
	// CloudNotifierStatusRunning ...
	CloudNotifierStatusRunning = "working"
	// CloudNotifierStatusDone ...
	CloudNotifierStatusDone = "finished"
	// CloudNotifierStatusErrored ...
	CloudNotifierStatusErrored = "errored"
	// CloudNotifierStatusTimeout ...
	CloudNotifierStatusTimeout = "timeout"
)

// NewCloudNotifier return a new status notifier. CloudURL should be of form
// {alias}/{bucket name}/logs/{request index}
func NewCloudNotifier(aliasName, bucketName, prefix string) (*CloudNotifier, error) {

	minioClient, err := GetMinioClient()
	if err != nil {
		return nil, xerrors.Errorf("failed to get the minion client: %v", err)
	}

	cn := &CloudNotifier{
		Terminated:  false,
		ID:          RandString(16),
		AliasName:   aliasName,
		BucketName:  bucketName,
		Prefix:      prefix,
		TaskEventCh: make(chan TaskEvent, 100),
		NotifDoneCh: make(chan struct{}),
		IsClosed:    false,
		Done:        make(chan bool),
		Status:      CloudNotifierStatusCreated,
	}

	closed, err := cn.isAlreadyDone()
	if err != nil {
		cn.Status = CloudNotifierStatusErrored
		return nil, errors.New("failed to check if already closed: " + err.Error())
	}

	if !closed {
		// With the first request, the bucket might not be created yet when the
		// cloud notifier is called. This is why we try 10 times to check if the
		// bucket exist before abandoning or starting to listen.
		retry := 10
		for retry > 0 {
			retry--
			found, err := minioClient.BucketExists(bucketName)
			if err != nil {
				cn.Status = CloudNotifierStatusErrored
				return nil, fmt.Errorf("failed to check if bucket '%s' exist: %s", bucketName, err.Error())
			}
			if found {
				break
			}
			if retry == 0 {
				cn.Status = CloudNotifierStatusErrored
				return nil, fmt.Errorf("the bucket '%s' does not exist", bucketName)
			}
			log.LLvlf1("bucket not found, trying again in 7 secs, remaaining tries: %d", retry)
			time.Sleep(time.Second * 7)
			retry--
		}
		cn.startListening()
	} else {
		log.LLvl1("cloud notifier is in closed state, thus not starting to listen")
		cn.close()
	}

	return cn, nil
}

// close terminates the sending of new notification and proprely acknowledges
// the receiver. Sets the status to 'done'
func (c *CloudNotifier) close() {
	c.closeStatus(CloudNotifierStatusDone)
}

// close terminates the sending of new notification and proprely acknowledges
// the receiver. Sets the status to 'errored'
func (c *CloudNotifier) closeError() {
	c.closeStatus(CloudNotifierStatusErrored)
}

// close terminates the sending of new notification and proprely acknowledges
// the receiver. Sets the status to 'timeOut'
func (c *CloudNotifier) closeTimeout() {
	c.closeStatus(CloudNotifierStatusTimeout)
}

// close terminates the sending of new notification and proprely acknowledges
// the receiver.
func (c *CloudNotifier) closeStatus(status CloudNotifierStatus) {
	c.Lock()
	c.Status = status
	c.IsClosed = true
	close(c.Done)
	close(c.NotifDoneCh)
	c.Unlock()
}

// isAlreadyDone checks if the last entry already stored on the cloud is of
// type 'closeOK' or 'closeError'. (If that's the case, then we won't wait for
// more entries to come.)
func (c *CloudNotifier) isAlreadyDone() (bool, error) {
	tasks, _, err := GetLogs(c.AliasName, c.BucketName, c.Prefix)
	if err != nil {
		return false, errors.New("failed to check if the cloud notifier is already " +
			"closed: " + err.Error())
	}
	if len(tasks) == 0 {
		return false, nil
	}
	if tasks[0].Type == TypeCloseError ||
		tasks[0].Type == TypeCloseOK {
		return true, nil
	}
	return false, nil
}

// startListening starts sending new taskEvent to the taskEventCh in a go
// routine.
func (c *CloudNotifier) startListening() {
	minioClient, err := GetMinioClient()
	if err != nil {
		log.Fatal("failed to get the minion client:", err)
	}

	c.Status = CloudNotifierStatusRunning
	go func() {
		// lets get our Task Event Factory (TEF)
		tef := NewTaskEventFactory("ds manager")
		// Listen for bucket notifications on "c.BucketName" filtered by prefix,
		// suffix and events.
		notifCh := minioClient.ListenBucketNotification(c.BucketName, c.Prefix,
			".json", []string{"s3:ObjectCreated:*"}, c.NotifDoneCh)
		for {
			select {
			case notifInfo := <-notifCh:
				if notifInfo.Err != nil {
					c.TaskEventCh <- tef.NewTaskEventError("failed to read notification",
						notifInfo.Err.Error())
					c.closeError()
					return
				}
				if len(notifInfo.Records) != 1 {
					c.TaskEventCh <- tef.NewTaskEventError("got an unexpected number "+
						"of records in notification",
						fmt.Sprintf("found %d notifications", len(notifInfo.Records)))
					continue
				}

				objectKey := notifInfo.Records[0].S3.Object.Key
				objectKey, err := url.QueryUnescape(objectKey)
				if err != nil {
					log.Error("failed to decode object key: " + err.Error())
					c.TaskEventCh <- tef.NewTaskEventError("failed to decode object key", err.Error())
					continue
				}
				object, err := minioClient.GetObject(c.BucketName, objectKey,
					minio.GetObjectOptions{})
				if err != nil {
					log.Error("error reading object: " + err.Error())
					c.TaskEventCh <- tef.NewTaskEventError("error reading object", err.Error())
					continue
				}

				logBuf, err := ioutil.ReadAll(object)
				if err != nil {
					log.Error("failed to copy the content of object: " + err.Error())
					c.TaskEventCh <- tef.NewTaskEventError("failed to copy the content of object", err.Error())
					continue
				}
				taskEvent := TaskEvent{}
				err = json.Unmarshal(logBuf, &taskEvent)
				if err != nil {
					log.Error("failed to unmarshal json log: " + err.Error())
					c.TaskEventCh <- tef.NewTaskEventError("failed to unmarshal json log", err.Error())
					continue
				}
				c.TaskEventCh <- taskEvent
				// Not intended to be used in a multi-threading environment
				if taskEvent.Type == TypeCloseError {
					c.closeError()
					return
				}
				if taskEvent.Type == TypeCloseOK {
					c.close()
					return
				}
			case <-time.After(time.Minute * 7):
				c.closeTimeout()
				return
			}
		}
	}()
}

// GetLogs returns the list of existing logs at given prefix. The list is sorted
// by the time of task events, with the most recent at the first index.
func GetLogs(aliasName, bucketName, prefix string) ([]*TaskEvent, CloudNotifierStatus, error) {
	minioClient, err := GetMinioClient()
	if err != nil {
		return nil, "", xerrors.Errorf("failed to get the minion client: %v", err)
	}

	found, err := minioClient.BucketExists(bucketName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check if bucket '%s' exist: %s", bucketName, err.Error())
	}
	tasks := make([]*TaskEvent, 0)
	if !found {
		log.LLvl1("bucket", bucketName, "not found, nothing to check")
		return tasks, "", nil
	}

	// Get the log objects
	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := false
	objectCh := minioClient.ListObjectsV2(bucketName, prefix+"/", isRecursive, doneCh)
	for objectInfo := range objectCh {
		if objectInfo.Err != nil {
			log.Error("error reading object info: " + objectInfo.Err.Error())
			continue
		}
		object, err := minioClient.GetObject(bucketName, objectInfo.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Error("error reading object: " + err.Error())
			continue
		}

		logBuf, err := ioutil.ReadAll(object)
		if err != nil {
			log.Error("failed to copy the content of object: " + err.Error())
			continue
		}

		taskEvent := TaskEvent{}
		err = json.Unmarshal(logBuf, &taskEvent)
		if err != nil {
			log.Errorf("failed to unmarshal json log '%s': %s", logBuf, err.Error())
			continue
		}

		tasks = append(tasks, &taskEvent)
	}
	sort.Sort(sort.Reverse(TaskEventSorter(tasks)))
	var status CloudNotifierStatus
	status = CloudNotifierStatusTimeout
	if len(tasks) > 0 {
		status = CloudNotifierStatusCreated
		switch tasks[0].Type {
		case "closeOK":
			status = CloudNotifierStatusDone
		case "closeError":
			status = CloudNotifierStatusDone
		case "info", "importantInfo":
			status = CloudNotifierStatusRunning
		}
	}
	return tasks, status, nil
}
