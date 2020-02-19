package projectc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dedis/odyssey/catalogc"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// The Project contract is used to express a query from a data scientist
// manager.

// ContractProjectID denotes a contract that can describe a project
var ContractProjectID = "odysseyproject"

// eachLine matches the content of non-empty lines
var eachLine = regexp.MustCompile(`(?m)^(.+)$`)

type contractProject struct {
	byzcoin.BasicContract
	ProjectData
}

// ProjectStatus describes the status of the project. It matches what you can
// find in dsmanager/app/models/projects.go
type ProjectStatus int

const (
	empty ProjectStatus = iota
	initialized

	preparing
	preparedOK
	preparedErrored

	updatingAttr
	updatedAttrOK
	updatedAttrErrored

	unlocking
	unlockedOK
	unlockedErrored

	deleting
	deletedOK
	deletedErrored
)

// ProjectData hold the data of the Odyssey project contract
type ProjectData struct {
	Datasets      []byzcoin.InstanceID
	Metadata      *catalogc.Metadata
	AccessPubKey  string
	EnclavePubKey string
	Status        ProjectStatus
	EnclaveURL    string
}

func (status ProjectStatus) String() string {
	statuses := []string{
		"empty",
		"initialized",
		"preparing",
		"preparedOK",
		"preparedErrored",
		"updatingAttr",
		"updatedAttrOK",
		"updatedAttrErrored",
		"unlocking",
		"unlockedOK",
		"unlockedErrored",
		"deleting",
		"deletedOK",
		"deletedErrored",
	}
	if status < empty || status > deletedErrored {
		return "NOT DEFINED"
	}

	return statuses[status]
}

// StatusFromString return a status given its string representation
func StatusFromString(status string) (ProjectStatus, error) {
	switch status {
	case "empty":
		return empty, nil
	case "initialized":
		return initialized, nil
	case "preparing":
		return preparing, nil
	case "preparedOK":
		return preparedOK, nil
	case "preparedErrored":
		return preparedErrored, nil
	case "updatingAttr":
		return updatingAttr, nil
	case "updatedAttrOK":
		return updatedAttrOK, nil
	case "updatedAttrErrored":
		return updatedAttrErrored, nil
	case "unlocking":
		return unlocking, nil
	case "unlockedOK":
		return unlockedOK, nil
	case "unlockedErrored":
		return unlockedErrored, nil
	case "deleting":
		return deleting, nil
	case "deletedOK":
		return deletedOK, nil
	case "deletedErrored":
		return deletedErrored, nil
	default:
		return -1, xerrors.Errorf("can't convert unkown status '%s'", status)
	}
}

// String returns a human readable string representation of the project data
func (pd ProjectData) String() string {
	out := new(strings.Builder)
	out.WriteString("- Project:\n")
	out.WriteString("-- Datasets:\n")
	for _, dataset := range pd.Datasets {
		fmt.Fprintf(out, "--- %x\n", dataset.Slice())
	}
	out.WriteString("-- Access pub key:\n")
	fmt.Fprintf(out, "--- %s\n", pd.AccessPubKey)
	out.WriteString("-- Enclave pub key:\n")
	fmt.Fprintf(out, "--- %s\n", pd.EnclavePubKey)
	out.WriteString("-- Status:\n")
	fmt.Fprintf(out, "--- %s\n", pd.Status)
	out.WriteString("-- Enclave URL:\n")
	fmt.Fprintf(out, "--- %s\n", pd.EnclaveURL)
	out.WriteString("-- Metadata:\n")
	if pd.Metadata != nil {
		out.WriteString(eachLine.ReplaceAllString(pd.Metadata.String(), "--$1"))
	}
	return out.String()
}

func contractProjectFromBytes(in []byte) (byzcoin.Contract, error) {
	cp := &contractProject{}
	err := protobuf.Decode(in, &cp.ProjectData)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// Spawning a new project contract is done when a data scientist has selected a
// list of datasets that he wants to use. The spawn then holds at this time a
// list of requested datasets.
func (c *contractProject) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction,
	coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	// Read and parse the project data, which normally contains only the list of
	// datasets, but user can actually fill it with more information.
	// We expect the following:
	// - instids: the list of instance ids (calypso write), a string separated by
	// 						comas
	// - pubkey: the public key that will have access to the enclave

	instID := inst.Spawn.Args.Search("datasetIDs")
	if instID == nil {
		return nil, nil, xerrors.New("datasetIDs not found in spawm arguments")
	}

	instIDstr := string(instID)

	r, err := regexp.Compile("^[0-9a-f]{64}(,[0-9a-f]{64})*$")
	if err != nil {
		return nil, nil, xerrors.New("failed to build regex: " + err.Error())
	}
	instIDstr = strings.Trim(instIDstr, " \n\r")
	ok := r.MatchString(instIDstr)
	if !ok {
		return nil, nil, xerrors.New("Got unexpected 'instids': " + instIDstr)
	}

	instIDlist := strings.Split(instIDstr, ",")

	pubKey := inst.Spawn.Args.Search("accessPubKey")
	if pubKey == nil {
		return nil, nil, xerrors.New("accessPubKey not found in spawn arguments")
	}

	pubKeyStr := string(pubKey)

	datasets := make([]byzcoin.InstanceID, len(instIDlist))
	for i, instID := range instIDlist {
		instidbuf, err := hex.DecodeString(instID)
		if err != nil {
			return nil, nil, xerrors.New("failed to decode instance id: " + err.Error())
		}
		datasets[i] = byzcoin.NewInstanceID(instidbuf)
	}

	projectData := ProjectData{}

	projectData.Status = empty
	projectData.AccessPubKey = pubKeyStr
	projectData.Datasets = datasets
	projectData.Metadata = &catalogc.Metadata{}
	projectDataBuf, err := protobuf.Encode(&projectData)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to encode back the projectData: %v", err)
	}

	sc = append(sc, byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""),
		ContractProjectID, projectDataBuf, darcID))
	return
}

func (c *contractProject) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction,
	coins []byzcoin.Coin) ([]byzcoin.StateChange, []byzcoin.Coin, error) {
	cout := coins
	var darcID darc.ID
	_, _, _, darcID, err := rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get darc: %v", err)
	}

	switch inst.Invoke.Command {
	case "update":
		projectDataBuf := inst.Invoke.Args.Search("projectData")
		var projectData ProjectData
		if len(projectDataBuf) == 0 {
			return nil, nil, xerrors.New("didn't find the 'projectData' argument")
		}
		err = protobuf.Decode(projectDataBuf, &projectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to decode projectData: %v", err)
		}

		projectDataBuf, err = protobuf.Encode(&projectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode back the projectData: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, projectDataBuf, darcID),
		}
		return sc, cout, nil
	case "updateStatus":
		statusBuf := inst.Invoke.Args.Search("status")
		if len(statusBuf) == 0 {
			return nil, nil, xerrors.New("didn't find the 'status' argument")
		}
		newStatus, err := StatusFromString(string(statusBuf))
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to get status from string: %v", err)
		}
		c.Status = newStatus
		projectDataBuf, err := protobuf.Encode(&c.ProjectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode project data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, projectDataBuf, darcID),
		}
		return sc, cout, nil
	case "updateMetadata":
		metadataJSONBuf := inst.Invoke.Args.Search("metadataJSON")
		if len(metadataJSONBuf) == 0 {
			return nil, nil, xerrors.Errorf("'metadataJSON' argument not found or empty")
		}

		metadata := &catalogc.Metadata{}
		err := json.Unmarshal(metadataJSONBuf, metadata)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to decode metadata from JSON: %v", err)
		}

		log.LLvl2("here is the metadata:", metadata)

		c.Metadata = metadata

		ProjectDataBuf, err := protobuf.Encode(&c.ProjectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, ProjectDataBuf, darcID),
		}
		return sc, cout, nil
	case "setURL":
		enclaveURLBuf := inst.Invoke.Args.Search("enclaveURL")
		if len(enclaveURLBuf) == 0 {
			return nil, nil, errors.New("didn't find the 'enclaveURL' argument")
		}
		c.EnclaveURL = string(enclaveURLBuf)
		projectDataBuf, err := protobuf.Encode(&c.ProjectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode project data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, projectDataBuf, darcID),
		}
		return sc, cout, nil
	case "setAccessPubKey":
		pubKeyBuff := inst.Invoke.Args.Search("pubKey")
		if len(pubKeyBuff) == 0 {
			return nil, nil, xerrors.New("didn't find the 'pubKey' argument")
		}
		c.AccessPubKey = string(pubKeyBuff)
		projectDataBuf, err := protobuf.Encode(&c.ProjectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode project data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, projectDataBuf, darcID),
		}
		return sc, cout, nil
	case "setEnclavePubKey":
		pubKeyBuff := inst.Invoke.Args.Search("pubKey")
		if len(pubKeyBuff) == 0 {
			return nil, nil, xerrors.New("didn't find the 'pubKey' argument")
		}
		c.EnclavePubKey = string(pubKeyBuff)
		projectDataBuf, err := protobuf.Encode(&c.ProjectData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode project data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractProjectID, projectDataBuf, darcID),
		}
		return sc, cout, nil
	default:
		return nil, nil, xerrors.Errorf("Unkown action '%s'", inst.Invoke.Command)
	}
}

// VerifyInstruction alows the one that has its key in the "EnclaveKey" field to
// do anything with the contract. This key should be the temporary key of the
// enclave.
func (c contractProject) VerifyInstruction(rst byzcoin.ReadOnlyStateTrie, instr byzcoin.Instruction, ctxHash []byte) error {

	// The enclave has the right to update the project instance. Here we check
	// if the identity used is the same as the one stored on the "enclavePubKey"
	// attribute of the contract.
	for i := range instr.Signatures {
		identity := instr.SignerIdentities[i]
		err := identity.Verify(ctxHash, instr.Signatures[i])
		if err == nil && identity.String() == c.EnclavePubKey {
			return nil
		}
	}

	return instr.Verify(rst, ctxHash)
}
