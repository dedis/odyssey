package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"

	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/controllers"
	"github.com/dedis/odyssey/enclavem/app/models"
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
)

// Flash ...
type Flash struct {
	i   interface{}
	Msg string
}

func main() {
	// Register the struct so encoding/gob knows about it
	gob.Register(helpers.Flash{})

	var err error
	conf, err = models.NewConfig()
	if err != nil {
		log.Fatal("failed to load config", err)
	}

	xlog.Info("loading db into memory")
	err = loadDb()
	if err != nil {
		log.Fatal("failed to import DB: " + err.Error())
	}

	flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		http.FileServer(http.Dir("."+"/assets/"))))
	router.HandleFunc("/favicon.ico", faviconHandler)

	router.Handle("/", http.HandlerFunc(controllers.HomeHandler(store, conf)))
	router.Handle("/healthz", healthz())
	router.Handle("/orgs", http.HandlerFunc(controllers.OrgsIndexHandler(store, conf)))
	router.Handle("/orgs/{id}", http.HandlerFunc(controllers.OrgsShowHandler(store, conf)))
	router.Handle("/vapps", http.HandlerFunc(controllers.VappsIndexHandler(store, conf)))
	router.Handle("/vapps/{id}", http.HandlerFunc(controllers.VappsShowHandler(store, conf)))
	router.Handle("/eprojects", http.HandlerFunc(controllers.EProjectsIndexHandler(store, conf)))
	router.Handle("/eprojects/{instID}/unlock", http.HandlerFunc(controllers.EProjectsShowUnlockHandler(store, conf)))
	router.Handle("/eprojects/{instID}", http.HandlerFunc(controllers.EProjectsShowHandler(store, conf)))
	router.Handle("/darcs", http.HandlerFunc(controllers.DarcIndexHandler(store, conf)))

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:        listenAddr,
		Handler:     tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:    logger,
		ReadTimeout: 500 * time.Second,
		// Important to have enought in order to poll the status of a task,
		// which can take some time when spawning an enclave.
		WriteTimeout: 600 * time.Second,
		// IdleTimeout:  15 * time.Second,
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

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
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

// parseConfig parses the config file and return a config struct
func parseConfig() (*models.Config, error) {
	conf := &models.Config{}
	_, err := toml.DecodeFile("config.toml", conf)
	if err != nil {
		return nil, errors.New("failed to read config: " + err.Error())
	}

	return conf, nil
}

func loadDb() error {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return errors.New("failed to open DB: " + err.Error())
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("EProjects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to create 'EProjects' bucket: " + err.Error())
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("EProjects"))
		b.ForEach(func(k, v []byte) error {
			project := &models.EProject{}
			err := json.Unmarshal(v, project)
			if err != nil {
				return errors.New("failed to unmarshal eproject: " + err.Error())
			}
			models.EProjectList[string(k)] = project
			return nil
		})
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
		err := tx.DeleteBucket([]byte("EProjects"))
		if err != nil {
			return errors.New("failed to delete the EProjects bucket: " + err.Error())
		}
		b, err := tx.CreateBucketIfNotExists([]byte("EProjects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		for k, project := range models.EProjectList {
			projectBuf, err := json.Marshal(project)
			if err != nil {
				return errors.New("failed to marshal eproject: " + err.Error())
			}
			err = b.Put([]byte(k), projectBuf)
			if err != nil {
				return errors.New("failed to save eproject buf: " + err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return errors.New("failed to update 'EProjects' bucket: " + err.Error())
	}
	return nil
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/images/favicon.ico")
}
