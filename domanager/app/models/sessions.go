package models

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
	"go.dedis.ch/onet/v3/log"
	"golang.org/x/xerrors"
)

// SessionsMap holds the session. It is public so that we can save and restore
// them from the main.go
var SessionsMap = map[string]*Session{}

// Session ...
type Session struct {
	BcPath string
	Cfg    *lib.Config
}

// GetSession return the current session. An error can happen if we can't get
// the session. If the session is not created we return an empty session.
func GetSession(store sessions.Store, r *http.Request) (*Session, error) {
	session, err := store.Get(r, "signin-session")
	if err != nil {
		return emptySession(), xerrors.Errorf("failed to get the signin "+
			"session: %v", err)
	}
	key, found := session.Values["key"]
	if !found {
		return emptySession(), nil
	}

	sess, found := SessionsMap[key.(string)]
	if !found {
		return emptySession(), nil
	}
	return sess, nil
}

// SaveSession saved the session to the static map
func SaveSession(ID string, bcPath string, cfg *lib.Config) {
	SessionsMap[ID] = &Session{bcPath, cfg}
}

func emptySession() *Session {
	return &Session{"", nil}
}

// IsLogged tells if the credential have been set
func (s Session) IsLogged() bool {
	return s.BcPath != "" && s.Cfg != nil
}

// LogIn sets the bcPath and cfg static variables by loading the config
func LogIn(bcPath string, store sessions.Store, r *http.Request, w http.ResponseWriter) error {
	cfg, _, err := lib.LoadConfig(bcPath)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	sess, err := store.Get(r, "signin-session")
	if err != nil {
		return xerrors.Errorf("failed to create signin-session: %v", err)
	}

	randKey := lib.RandString(12)
	sess.Values["key"] = randKey

	SaveSession(randKey, bcPath, &cfg)

	err = sess.Save(r, w)
	if err != nil {
		return xerrors.Errorf("failed to save session: %v", err)
	}

	return nil
}

// GetRoster ...
func (s Session) GetRoster() string {
	return fmt.Sprintf("Roster:\n%+v", s.Cfg.Roster)
}

// GetByzCoinID ...
func (s Session) GetByzCoinID() string {
	return fmt.Sprintf("%x", s.Cfg.ByzCoinID)
}

// GetAdminDarc ...
func (s Session) GetAdminDarc() string {
	return s.Cfg.AdminDarc.String()
}

// GetIdentity ...
func (s Session) GetIdentity() string {
	return s.Cfg.AdminIdentity.String()
}

// GetShortIdentity return the identity stored in the config in a short format
func (s Session) GetShortIdentity() string {
	if len(s.Cfg.AdminIdentity.String()) > 13 {
		return fmt.Sprintf("%.12s...", s.Cfg.AdminIdentity.String())
	}
	return s.Cfg.AdminIdentity.String()
}

// PrepareAfterUnmarshal should be called after loading a savec Session
func (s *Session) PrepareAfterUnmarshal() error {
	cfg, _, err := lib.LoadConfig(s.BcPath)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}
	s.Cfg = &cfg

	return nil
}

// PrepareBeforeMarshal should be called before the session is marshalled
func (s *Session) PrepareBeforeMarshal() {
	s.Cfg = nil
}

// Destroy removes the session from the maps and delets the config file
func (s Session) Destroy(store sessions.Store, r *http.Request,
	w http.ResponseWriter) error {

	sess, err := store.Get(r, "signin-session")
	if err != nil {
		return xerrors.Errorf("failed to create signin-session: %v", err)
	}

	key, found := sess.Values["key"]
	if !found {
		return xerrors.Errorf("key value not found in the session")
	}

	_, found = SessionsMap[key.(string)]
	if !found {
		return xerrors.Errorf("session with key '%s' not found in the map", key.(string))
	}

	log.Info("deleting the session with key", key.(string))
	delete(SessionsMap, key.(string))
	delete(sess.Values, key)

	sess.Save(r, w)
	log.Info("session deleted from the map and the session variable")

	return nil
}
