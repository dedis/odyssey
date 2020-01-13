package catalogc

import (
	"encoding/json"
	"regexp"

	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// eachLine matches the content of non-empty lines
var eachLine = regexp.MustCompile(`(?m)^(.+)$`)

// The catalog contract is used to keep track of the owners and their datasets
// that are stored along with their metadata.

// ContractCatalogID denotes a contract that can store a catalog of owners and
// datasets
var ContractCatalogID = "odysseycatalog"

type contractCatalog struct {
	byzcoin.BasicContract
	CatalogData
}

func contractCatalogFromBytes(in []byte) (byzcoin.Contract, error) {
	cc := &contractCatalog{}
	err := protobuf.Decode(in, &cc.CatalogData)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// Spawn initializes the fields of the catalog data Note: apparently our
// protobuf doesn't support an initialized map, it is still nill even after
// initialization. For this reason we will perform lazy initialization a be sure
// to check before accessing a potentially nil attribute.
func (c *contractCatalog) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction,
	coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	catalogData := CatalogData{}
	catalogDataBuf, err := protobuf.Encode(&catalogData)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
	}

	sc = append(sc, byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""),
		ContractCatalogID, catalogDataBuf, darcID))
	return
}

func (c *contractCatalog) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction,
	coins []byzcoin.Coin) ([]byzcoin.StateChange, []byzcoin.Coin, error) {
	cout := coins
	var darcID darc.ID
	_, _, _, darcID, err := rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get darc: %v", err)
	}

	switch inst.Invoke.Command {
	case "addOwner":
		firstnameBuf := inst.Invoke.Args.Search("firstname")
		if len(firstnameBuf) == 0 {
			return nil, nil, xerrors.Errorf("'firstname' argument is empty")
		}
		fistname := string(firstnameBuf)

		lastnameBuf := inst.Invoke.Args.Search("lastname")
		if len(lastnameBuf) == 0 {
			return nil, nil, xerrors.Errorf("'lastname' argument is empty")
		}
		lastname := string(lastnameBuf)

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument is empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner != nil {
			return nil, nil, xerrors.Errorf("owner with identity string '%s' "+
				"already exist in the catalog:\n%s", identityStr, owner.String())
		}

		newOwner := &Owner{}
		newOwner.Firstname = fistname
		newOwner.Lastname = lastname
		newOwner.Datasets = make([]*Dataset, 0)
		newOwner.IdentityStr = identityStr

		err = c.AddOwner(newOwner)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to add owner: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil
	case "updateOwner":
		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the list of owners is nil, nothing to update then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStrBuf' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		firstnameBuf := inst.Invoke.Args.Search("firstname")
		if len(firstnameBuf) != 0 {
			owner.Firstname = string(firstnameBuf)
		}

		lastnameBuf := inst.Invoke.Args.Search("lastname")
		if len(lastnameBuf) != 0 {
			owner.Lastname = string(lastnameBuf)
		}

		newIdentityStrBuff := inst.Invoke.Args.Search("newIdentityStr")
		if len(newIdentityStrBuff) != 0 {
			owner.IdentityStr = string(newIdentityStrBuff)
			for _, dataset := range owner.Datasets {
				dataset.IdentityStr = string(newIdentityStrBuff)
			}
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil
	case "deleteOwner":
		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the list of owners is nil, nothing to delete then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		err = c.RemoveOwner(identityStr)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to remove owner: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil
	case "addDataset":
		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the map of owners is nil, nothing to add then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		calypsoWriteIDBuff := inst.Invoke.Args.Search("calypsoWriteID")
		if len(calypsoWriteIDBuff) == 0 {
			return nil, nil, xerrors.Errorf("'calypsoWriteID' argument not found or empty")
		}
		calypsoWriteID := string(calypsoWriteIDBuff)

		dataset := owner.GetDataset(calypsoWriteID)
		if dataset != nil {
			return nil, nil, xerrors.Errorf("Dataset '%s' already exist:\n%s",
				calypsoWriteID, dataset)
		}

		datasetBuf := inst.Invoke.Args.Search("dataset")
		if len(datasetBuf) == 0 {
			return nil, nil, xerrors.Errorf("'dataset' argument not found or empty")
		}

		dataset = &Dataset{}
		err = protobuf.Decode(datasetBuf, dataset)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to decode dataset: %v", err)
		}

		dataset.CalypsoWriteID = calypsoWriteID
		dataset.IdentityStr = identityStr
		dataset.Metadata = c.CatalogData.Metadata
		// By default a dataset is not in an "archived" state
		dataset.IsArchived = false

		err = owner.AddDataset(dataset)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to add dataset: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil
	case "updateDataset":
		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the map of owners is nil, nothing to update then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		calypsoWriteIDBuff := inst.Invoke.Args.Search("calypsoWriteID")
		if len(calypsoWriteIDBuff) == 0 {
			return nil, nil, xerrors.Errorf("'calypsoWriteID' argument not found or empty")
		}
		calypsoWriteID := string(calypsoWriteIDBuff)

		newCalypsoWriteIDBuff := inst.Invoke.Args.Search("newCalypsoWriteID")
		newCalypsoWriteID := string(newCalypsoWriteIDBuff)

		foundDataset := owner.GetDataset(calypsoWriteID)
		if foundDataset == nil {
			return nil, nil, xerrors.Errorf("Can't update, dataset '%s' not found",
				calypsoWriteID)
		}

		datasetBuf := inst.Invoke.Args.Search("dataset")
		if len(datasetBuf) == 0 {
			return nil, nil, xerrors.Errorf("'dataset' argument not found or empty")
		}

		dataset := &Dataset{}
		err = protobuf.Decode(datasetBuf, dataset)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to decode dataset: %v", err)
		}

		dataset.IdentityStr = identityStr

		if newCalypsoWriteID == "" {
			dataset.CalypsoWriteID = calypsoWriteID
		} else {
			dataset.CalypsoWriteID = newCalypsoWriteID
		}

		err = owner.ReplaceDataset(calypsoWriteID, dataset)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to replace dataset: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil

	case "archiveDataset":
		// Here we set the IsArchived attribute on the dataset and remove all
		// its attributes (ie. set an empty Metadata field)

		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the map of owners is nil, nothing to update then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		calypsoWriteIDBuff := inst.Invoke.Args.Search("calypsoWriteID")
		if len(calypsoWriteIDBuff) == 0 {
			return nil, nil, xerrors.Errorf("'calypsoWriteID' argument not found or empty")
		}
		calypsoWriteID := string(calypsoWriteIDBuff)

		foundDataset := owner.GetDataset(calypsoWriteID)
		if foundDataset == nil {
			return nil, nil, xerrors.Errorf("Can't archive, dataset '%s' not found",
				calypsoWriteID)
		}

		err = owner.ArchiveDataset(calypsoWriteID)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to archive the dataset: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil

	case "deleteDataset":
		// Althought we use the term "delete", what we do here is only removing
		// it from the catalog, which doesn't swipe it entirely from the ledger.
		// This is more a debug feature because we sometime want to remove a
		// dataset during developpment.

		if c.Owners == nil {
			return nil, nil, xerrors.Errorf("the map of owners is nil, nothing to update then")
		}

		identityStrBuf := inst.Invoke.Args.Search("identityStr")
		if len(identityStrBuf) == 0 {
			return nil, nil, xerrors.Errorf("'identityStr' argument not found or empty")
		}
		identityStr := string(identityStrBuf)

		owner := c.GetOwner(identityStr)
		if owner == nil {
			return nil, nil, xerrors.Errorf("owner '%s' not found", identityStr)
		}

		calypsoWriteIDBuff := inst.Invoke.Args.Search("calypsoWriteID")
		if len(calypsoWriteIDBuff) == 0 {
			return nil, nil, xerrors.Errorf("'calypsoWriteID' argument not found or empty")
		}
		calypsoWriteID := string(calypsoWriteIDBuff)

		foundDataset := owner.GetDataset(calypsoWriteID)
		if foundDataset == nil {
			return nil, nil, xerrors.Errorf("Can't delete, dataset '%s' not found",
				calypsoWriteID)
		}

		err = owner.DeleteDataset(calypsoWriteID)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to remove the dataset: %v", err)
		}

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil

	case "updateMetadata":
		metadataJSONBuf := inst.Invoke.Args.Search("metadataJSON")
		if len(metadataJSONBuf) == 0 {
			return nil, nil, xerrors.Errorf("'metadataJSON' argument not found or empty")
		}

		metadata := &Metadata{}
		err := json.Unmarshal(metadataJSONBuf, metadata)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to decode metadata from JSON: %v", err)
		}

		c.Metadata = metadata

		catalogDataBuf, err := protobuf.Encode(&c.CatalogData)
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to encode the catalog data: %v", err)
		}
		sc := []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCatalogID, catalogDataBuf, darcID),
		}
		return sc, cout, nil
	default:
		return nil, nil, xerrors.Errorf("Unkown action '%s'", inst.Invoke.Command)
	}
}

// VerifyInstruction allows an owner to add and update a dataset. The owner must
// be added before by someone that has the invoke:odysseycatalog.addOwner right.
func (c contractCatalog) VerifyInstruction(rst byzcoin.ReadOnlyStateTrie,
	instr byzcoin.Instruction, ctxHash []byte) error {

	if instr.Action() == "invoke:odysseycatalog.addDataset" ||
		instr.Action() == "invoke:odysseycatalog.updateDataset" ||
		instr.Action() == "invoke:odysseycatalog.deleteDataset" ||
		instr.Action() == "invoke:odysseycatalog.archiveDataset" {
		identityStrBuf := instr.Invoke.Args.Search("identityStr")
		identityStr := string(identityStrBuf)

		for i := range instr.Signatures {
			identity := instr.SignerIdentities[i]
			err := identity.Verify(ctxHash, instr.Signatures[i])
			if err == nil && identityStr == identity.String() {
				return nil
			}
		}
	}

	return instr.Verify(rst, ctxHash)
}
