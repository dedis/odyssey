package helpers

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// FlashType ...
type FlashType int

const (
	// Info ...
	Info FlashType = iota
	// Warning ...
	Warning
	// Error ...
	Error
)

// Flash ...
type Flash struct {
	Msg  string
	Type FlashType
}

// AddFlash ...
func AddFlash(w http.ResponseWriter, r *http.Request, msg string, store sessions.Store, ft FlashType) error {
	session, err := store.Get(r, "flash-session")
	if err != nil {
		return errors.New("failed to get flash-session: " + err.Error())
	}
	session.AddFlash(Flash{Msg: msg, Type: ft})
	err = session.Save(r, w)
	if err != nil {
		return errors.New("failed to save session: " + err.Error())
	}
	return nil
}

// ExtractFlash ...
func ExtractFlash(w http.ResponseWriter, r *http.Request, store sessions.Store) ([]Flash, error) {
	session, err := store.Get(r, "flash-session")
	if err != nil {
		return nil, errors.New("failed to get flash-session: " + err.Error())
	}
	flashes := session.Flashes()
	res := make([]Flash, len(flashes))
	for i, flash := range flashes {
		res[i] = flash.(Flash)
	}
	session.Save(r, w)
	return res, nil
}
