package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync/atomic"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dedis/odyssey/domanager/app/controllers"
	"github.com/dedis/odyssey/domanager/app/models"
	dsmanagercontrollers "github.com/dedis/odyssey/dsmanager/app/controllers"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	xlog "go.dedis.ch/onet/v3/log"
	bolt "go.etcd.io/bbolt"
)

type key int

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	healthy    int32
	store      = sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf       *models.Config
	db         *bolt.DB
)

func init() {
	var err error
	conf, err = parseConfig()
	if err != nil {
		log.Fatal("failed to load config", err)
	}
}

// @title Data Scientist Manager REST API
// @version 1.0
// @description REST functionalities provided by the Data Scientist Manager

// @host localhost:5001
// @BasePath /v2
func main() {
	// Register the struct so encoding/gob knows about it
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	xlog.LLvl1("here is the catalog id:", conf.CatalogID)

	xlog.Info("loading db into memory")
	err := loadDb()
	if err != nil {
		log.Fatal("failed to import DB: " + err.Error())
	}

	flag.StringVar(&listenAddr, "listen-addr", ":5002", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		http.FileServer(http.Dir("."+"/assets/"))))
	router.HandleFunc("/favicon.ico", faviconHandler)

	router.Handle("/signin", http.HandlerFunc(controllers.SessionHandler(store, conf)))
	router.Handle("/sessions", http.HandlerFunc(controllers.SessionHandler(store, conf)))
	router.Handle("/profile", http.HandlerFunc(controllers.SessionProfileHandler(store, conf)))
	router.Handle("/datasets", http.HandlerFunc(controllers.DatasetIndexHandler(store, conf)))
	router.Handle("/datasets/new", http.HandlerFunc(controllers.DatasetNewHandler(store, conf)))
	router.Handle("/datasets/{id}", http.HandlerFunc(controllers.DatasetShowHandler(store, conf)))
	router.Handle("/datasets/{id}/attributes", http.HandlerFunc(controllers.DatasetShowAttributesHandler(store, conf)))
	router.Handle("/datasets/{id}/archive", http.HandlerFunc(controllers.DatasetShowArchiveHandler(store, conf)))
	router.Handle("/datasets/{id}/audit", http.HandlerFunc(controllers.DatasetShowAuditHandler(store, conf)))
	router.Handle("/datasets/{id}/debug", http.HandlerFunc(controllers.DatasetShowDebugHandler(store, conf)))
	router.Handle("/", http.HandlerFunc(controllers.HomeHandler(store, conf)))
	router.Handle("/healthz", healthz())
	router.Handle("/showtasks", http.HandlerFunc(controllers.ShowtasksIndexHandler(store, conf)))
	router.Handle("/showtasks/{id}", http.HandlerFunc(controllers.ShowtasksShowHandler(store, conf)))
	router.Handle("/tasks/{id}", http.HandlerFunc(dsmanagercontrollers.TasksShowHandler(store)))
	router.Handle("/lifecycle", http.HandlerFunc(controllers.ShowLifecycle(store, conf)))

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  50 * time.Second,
		WriteTimeout: 600 * time.Second,
		// IdleTimeout:  150 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		logger.Println("Saving memory variables into DB...")
		err := saveDb()
		if err != nil {
			xlog.Error("failed to save the db: " + err.Error())
		}

		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

// parseConfig parses the config file and return a config struct
func parseConfig() (*models.Config, error) {
	conf := &models.Config{}
	_, err := toml.DecodeFile("config.toml", conf)
	if err != nil {
		return nil, errors.New("failed to read config: " + err.Error())
	}

	return conf, nil
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/images/favicon.ico")
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	value := 12
	p := &struct{ Value int }{value}

	// cl := calypso.NewClient(nil)

	t, _ := template.ParseFiles("test.html")
	t.Execute(w, p)
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loadDb() error {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return errors.New("failed to open DB: " + err.Error())
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Sessions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to create 'Sessions' bucket: " + err.Error())
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		b.ForEach(func(k, v []byte) error {
			session := &models.Session{}
			err := json.Unmarshal(v, session)
			if err != nil {
				xlog.Error("failed to unmarshal session: " + err.Error())
				return errors.New("failed to unmarshal session: " + err.Error())
			}
			err = session.PrepareAfterUnmarshal()
			if err != nil {
				xlog.Error("failed to prepare after marshal: " + err.Error())
				return errors.New("failed to prepare after marshal: " + err.Error())
			}

			models.SessionsMap[string(k)] = session
			return nil
		})
		return nil
	})

	// Tasks

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to create 'Tasks' bucket: " + err.Error())
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		tempTaskList := make([]*xhelpers.Task, 0)
		b.ForEach(func(k, v []byte) error {
			task := &xhelpers.Task{}
			err := json.Unmarshal(v, task)
			if err != nil {
				xlog.Error("failed to unmarshal session: " + err.Error())
				return errors.New("failed to unmarshal session: " + err.Error())
			}
			task.PrepareAfterUnmarshal()

			tempTaskList = append(tempTaskList, task)
			return nil
		})
		sort.Sort(xhelpers.TaskSorter(tempTaskList))
		xhelpers.TaskList = tempTaskList
		return nil
	})

	return nil
}

func saveDb() error {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return errors.New("failed to open DB: " + err.Error())
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("Sessions"))
		if err != nil {
			return errors.New("failed to delete the Sessions bucket: " + err.Error())
		}
		b, err := tx.CreateBucketIfNotExists([]byte("Sessions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		for k, session := range models.SessionsMap {
			session.PrepareBeforeMarshal()
			sessionBuf, err := json.Marshal(session)
			if err != nil {
				return errors.New("failed to marshal session: " + err.Error())
			}
			err = b.Put([]byte(k), sessionBuf)
			if err != nil {
				return errors.New("failed to save session buf: " + err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to update 'Sessions' bucket: " + err.Error())
	}

	// Task

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("Tasks"))
		if err != nil {
			return errors.New("failed to delete the Tasks bucket: " + err.Error())
		}
		b, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		for _, task := range xhelpers.TaskList {
			task.PrepareBeforeMarshal()
			taskBuf, err := json.Marshal(task)
			if err != nil {
				return errors.New("failed to marshal task: " + err.Error())
			}
			err = b.Put([]byte(task.ID), taskBuf)
			if err != nil {
				return errors.New("failed to save task buf: " + err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to update 'Tasks' bucket: " + err.Error())
	}
	return nil
}
