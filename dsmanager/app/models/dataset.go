package models

// Dataset contains the public informations of a dataset. This should match
// exactly what is stored as json in the extra data field of a write instance,
// as this struct is used to directly decode the json content.
type Dataset struct {
	Title       string
	Description string
	Author      string
	WriteInstID string
	CloudURL    string
	SHA2        string
}
