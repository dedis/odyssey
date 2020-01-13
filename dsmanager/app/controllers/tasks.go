package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
)

// TasksShowHandler ...
func TasksShowHandler(gs *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tasksShowGet(w, r, gs)
		default:
			log.Error("Only GET is supported for /tasks/{id}")
		}
	}
}

func tasksShowGet(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		helpers.RedirectWithErrorFlash("/", "the response writer does not "+
			"support streaming", w, r, store)
		return
	}
	tef := helpers.NewTaskEventFFactory("ds manager", flusher, w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	params := mux.Vars(r)
	indexStr := params["id"]
	if indexStr == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get param")
		flusher.Flush()
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		tef.FlushTaskEventCloseErrorf("wrong index", "failed to convert %s "+
			"to int: %s", indexStr, err.Error())
		return
	}

	if index >= len(helpers.TaskList) || index < 0 {
		tef.FlushTaskEventCloseErrorf("index out of bound", "Index out of "+
			"bound: 0 > (index) %d >= len(TaskList) %d", index,
			len(helpers.TaskList))
		return
	}
	task := helpers.TaskList[index]

	// If the task is not in a working state then there is nothing to return
	// here.
	var client *helpers.Subscriber
	if task.Status == helpers.StatusWorking {
		client = task.Subscribe()
	} else {
		log.LLvl1("task not in the working status, nothing to send")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var taskEl *helpers.TaskEvent
	// for _, taskEl = range client.PastEvents {
	// 	fmt.Fprint(w, taskEl.JSONStream())
	// 	flusher.Flush()
	// 	time.Sleep(time.Second * 1)
	// }

	for {
		select {
		case taskEl = <-client.TaskStream:
			fmt.Fprint(w, taskEl.JSONStream())
			flusher.Flush()
		default:
			select {
			case <-client.Done:
				// We ensure we empty our Tastream before exiting
			streamFlush:
				for {
					select {
					case taskEl = <-client.TaskStream:
						fmt.Fprint(w, taskEl.JSONStream())
						flusher.Flush()
					default:
						break streamFlush
					}
				}
				log.Info("client done...")
				return
			default:
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}
