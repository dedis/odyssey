package models

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"go.dedis.ch/onet/v3/log"
)

// EProjectList holds all the projects
var EProjectList = make(map[string]*EProject)

const (
	// EProjectStatusBootingEnclave is set when the enclave is booting
	EProjectStatusBootingEnclave = "bootingEnclave"
	// EProjectStatusSetupReady is set when the enclave is ready
	EProjectStatusSetupReady = "setupReady"
	// EProjectStatusSetupErrored is set when the enclave set up failed
	EProjectStatusSetupErrored = "setupErrored"
	// EProjectStatusUnlockingEnclave is set when trying to unlock the enclave
	EProjectStatusUnlockingEnclave = "unlockingEnclave"
	// EProjectStatusUnlockingEnclaveDone is set when the enclave is unlocked
	EProjectStatusUnlockingEnclaveDone = "unlockingEnclaveDone"
	// EProjectStatusUnlockingEnclaveErrored is set when unlocking the enclave
	// failed
	EProjectStatusUnlockingEnclaveErrored = "unlockingEnclaveErrored"
)

// EProjectStatus describes the status of a eproject
type EProjectStatus string

// EProject holds a project which is tied to a project instance ID and an
// enclave. "E" stands for "Enclave".
type EProject struct {
	// This ProjectUID is intended to be the same as the ProjectID in the
	// datascientist manager.
	ProjectUID    string
	InstanceID    string
	EnclaveHref   string
	EnclaveName   string
	Status        EProjectStatus
	IPAddr        net.IP
	PubKey        string
	CloudEndpoint string
	ReadInstIDs   []string
	WriteInstIDs  []string
}

// ParseKey extract the public key, which is stored in the <type>:<key> format
func (e EProject) ParseKey() (string, error) {
	keySlice := strings.Split(e.PubKey, ":")
	if len(keySlice) != 2 {
		return "", fmt.Errorf("failed to split the key '%s', "+
			"got a lenght of %d", e.PubKey, len(keySlice))
	}
	return keySlice[1], nil
}

// UpdateProjectcStatus updates the status of the project instance. This should
// be put in the helpers package but since it is using Config, it would create a
// dependence cycle.
func UpdateProjectcStatus(conf *Config, status string, instanceID string) error {

	cmd := exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "invoke", "updateStatus", "-bc", conf.BCPath, "-sign",
		conf.KeyID, "-status", status, "-i", instanceID)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	var err error
	retry := 4
	for retry > 0 {
		err = cmd.Run()
		if err == nil {
			break
		}
		retry--
		log.Warnf("failed to update the projectc status, %d try left, "+
			"sleeping for 7s...: %s", retry, err)
		time.Sleep(7 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to update the project instance status after 4 try, "+
			"failed to run the command: %s - "+
			"Output: %s - Err: %s", err.Error(), outb.String(), errb.String())
	}

	return nil
}
