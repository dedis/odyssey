package clicontracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/urfave/cli"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// CatalogSpawn spawns a new instance of a catalog contract
func CatalogSpawn(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	dstr := c.String("darc")
	if dstr == "" {
		dstr = cfg.AdminDarc.GetIdentityString()
	}
	d, err := lib.GetDarcByString(cl, dstr)
	if err != nil {
		return err
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(d.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: catalogc.ContractCatalogID,
			Args:       byzcoin.Arguments{},
		},
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return xerrors.Errorf("failed to create transaction: %v", err)
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	instID := ctx.Instructions[0].DeriveID("")
	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	proof, err := cl.WaitProof(instID, time.Second, nil)
	if err != nil {
		return errors.New("couldn't get proof: " + err.Error())
	}

	_, resultBuf, _, _, err := proof.KeyValue()
	if err != nil {
		return errors.New("couldn't get value out of proof: " + err.Error())
	}

	result := catalogc.CatalogData{}
	err = protobuf.Decode(resultBuf, &result)
	if err != nil {
		return errors.New("couldn't decode the result: " + err.Error())
	}

	iidStr := hex.EncodeToString(instID.Slice())
	log.Infof("Spawned a new catalog instance. "+
		"Its instance id is:\n%s\n%s", iidStr, result)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeAddOwner adds a new owner
func CatalogInvokeAddOwner(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	firstnameStr := c.String("firstname")
	if firstnameStr == "" {
		return errors.New("please provide the firstname with --firstname")
	}

	lastnameStr := c.String("lastname")
	if lastnameStr == "" {
		return errors.New("please provide the lastname with --lastname")
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "addOwner",
		Args: byzcoin.Arguments{
			{
				Name: "firstname", Value: []byte(firstnameStr),
			},
			{
				Name: "lastname", Value: []byte(lastnameStr),
			},
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Catalog contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeUpdateOwner updates the attributes of an owner
func CatalogInvokeUpdateOwner(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	firstnameStr := c.String("firstname")
	lastnameStr := c.String("lastname")
	newIdentityStr := c.String("newIdentityStr")

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "updateOwner",
		Args: byzcoin.Arguments{
			{
				Name: "firstname", Value: []byte(firstnameStr),
			},
			{
				Name: "lastname", Value: []byte(lastnameStr),
			},
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "newIdentityStr", Value: []byte(newIdentityStr),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Owner updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeDeleteOwner delete an owner
func CatalogInvokeDeleteOwner(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "deleteOwner",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Owner deleted! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeAddDataset delete an owner
func CatalogInvokeAddDataset(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	calypsoWriteID := c.String("calypsoWriteID")
	if calypsoWriteID == "" {
		return errors.New("please provide the calypsoWriteID with --calypsoWriteID")
	}

	title := c.String("title")
	if title == "" {
		return errors.New("please provide the title with --title")
	}

	description := c.String("description")
	if description == "" {
		return errors.New("please provide the description with --description")
	}

	cloudURL := c.String("cloudURL")
	if cloudURL == "" {
		return errors.New("please provide the cloudURL with --cloudURL")
	}

	sha2 := c.String("sha2")
	if sha2 == "" {
		return errors.New("please provide the sha2 with --sha2")
	}

	dataset := catalogc.Dataset{
		Title:       title,
		Description: description,
		CloudURL:    cloudURL,
		SHA2:        sha2,
	}

	datasetBuf, err := protobuf.Encode(&dataset)
	if err != nil {
		return xerrors.Errorf("failed to encode dataset: %v", err)
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "addDataset",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "calypsoWriteID", Value: []byte(calypsoWriteID),
			},
			{
				Name: "dataset", Value: datasetBuf,
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Dataset added! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeUpdateDataset delete an owner
func CatalogInvokeUpdateDataset(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	// We get the catalog data to update only what is given in argument
	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return xerrors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return xerrors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return xerrors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return xerrors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID,
		&catalogData)
	if err != nil {
		return xerrors.New("couldn't get a project instance: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	calypsoWriteID := c.String("calypsoWriteID")
	if calypsoWriteID == "" {
		return errors.New("please provide the calypsoWriteID with --calypsoWriteID")
	}

	if catalogData.Owners == nil {
		return xerrors.Errorf("CatalogData.Owners map is nil, nothing to update then")
	}
	owner := catalogData.GetOwner(identityStr)
	if owner == nil {
		return xerrors.Errorf("Owner with identity '%s' not found", identityStr)
	}

	if owner.Datasets == nil {
		return xerrors.Errorf("the map of datasets is nil, nothing to update then")
	}
	dataset := owner.GetDataset(calypsoWriteID)
	if dataset == nil {
		return xerrors.Errorf("dataset with id '%s' not found", calypsoWriteID)
	}

	title := c.String("title")
	if title == "_" {
		dataset.Title = ""
	} else if title != "" {
		dataset.Title = title
	}

	description := c.String("description")
	if description == "_" {
		dataset.Description = ""
	} else if description != "" {
		dataset.Description = description
	}

	cloudURL := c.String("cloudURL")
	if cloudURL == "_" {
		dataset.CloudURL = ""
	} else if cloudURL != "" {
		dataset.CloudURL = cloudURL
	}

	sha2 := c.String("sha2")
	if sha2 == "_" {
		dataset.SHA2 = ""
	} else if sha2 != "" {
		dataset.SHA2 = sha2
	}

	metadataJSON := c.String("metadataJSON")
	if metadataJSON == "_" {
		dataset.Metadata = nil
	} else if metadataJSON != "" {
		metadataJSONBuf := []byte(metadataJSON)
		metadata := &catalogc.Metadata{}
		err := json.Unmarshal(metadataJSONBuf, metadata)
		if err != nil {
			return xerrors.Errorf("failed to decode metadata from JSON: %v", err)
		}
		dataset.Metadata = metadata
	}

	datasetBuf, err := protobuf.Encode(dataset)
	if err != nil {
		return xerrors.Errorf("failed to encode dataset: %v", err)
	}

	// if this one is empty, the contract will not use it
	newCalypsoWriteID := c.String("newCalypsoWriteID")

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "updateDataset",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "calypsoWriteID", Value: []byte(calypsoWriteID),
			},
			{
				Name: "newCalypsoWriteID", Value: []byte(newCalypsoWriteID),
			},
			{
				Name: "dataset", Value: datasetBuf,
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Dataset updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeArchiveDataset archives a dataset
func CatalogInvokeArchiveDataset(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	// We get the catalog data to update only what is given in argument
	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return xerrors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return xerrors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return xerrors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return xerrors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID,
		&catalogData)
	if err != nil {
		return xerrors.New("couldn't get a project instance: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	calypsoWriteID := c.String("calypsoWriteID")
	if calypsoWriteID == "" {
		return errors.New("please provide the calypsoWriteID with --calypsoWriteID")
	}

	if catalogData.Owners == nil {
		return xerrors.Errorf("CatalogData.Owners map is nil, nothing to delete then")
	}
	owner := catalogData.GetOwner(identityStr)
	if owner == nil {
		return xerrors.Errorf("Owner with identity '%s' not found", identityStr)
	}

	if owner.Datasets == nil {
		return xerrors.Errorf("the map of datasets is nil, nothing to delete then")
	}
	dataset := owner.GetDataset(calypsoWriteID)
	if dataset == nil {
		return xerrors.Errorf("dataset with id '%s' not found", calypsoWriteID)
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "archiveDataset",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "calypsoWriteID", Value: []byte(calypsoWriteID),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Dataset deleted! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeDeleteDataset delete a dataset
func CatalogInvokeDeleteDataset(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	// We get the catalog data to update only what is given in argument
	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return xerrors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return xerrors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return xerrors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return xerrors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID,
		&catalogData)
	if err != nil {
		return xerrors.New("couldn't get a project instance: " + err.Error())
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return errors.New("failed to parse the signer: " + err.Error())
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	calypsoWriteID := c.String("calypsoWriteID")
	if calypsoWriteID == "" {
		return errors.New("please provide the calypsoWriteID with --calypsoWriteID")
	}

	if catalogData.Owners == nil {
		return xerrors.Errorf("CatalogData.Owners map is nil, nothing to delete then")
	}
	owner := catalogData.GetOwner(identityStr)
	if owner == nil {
		return xerrors.Errorf("Owner with identity '%s' not found", identityStr)
	}

	if owner.Datasets == nil {
		return xerrors.Errorf("the map of datasets is nil, nothing to delete then")
	}
	dataset := owner.GetDataset(calypsoWriteID)
	if dataset == nil {
		return xerrors.Errorf("dataset with id '%s' not found", calypsoWriteID)
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "deleteDataset",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "calypsoWriteID", Value: []byte(calypsoWriteID),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Dataset deleted! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogInvokeUpdateMetadata upadte the metadata
func CatalogInvokeUpdateMetadata(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return xerrors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return xerrors.Errorf("failed to load config: %v", err)
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return xerrors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return xerrors.New("failed to decode the instid string: " + err.Error())
	}

	metadataJSON := c.String("metadataJSON")
	if metadataJSON == "" {
		return xerrors.New("--metadataJSON flag is required")
	}

	var signer *darc.Signer

	sstr := c.String("sign")
	if sstr == "" {
		signer, err = lib.LoadKey(cfg.AdminIdentity)
	} else {
		signer, err = lib.LoadKeyFromString(sstr)
	}
	if err != nil {
		return xerrors.New("failed to parse the signer: " + err.Error())
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: catalogc.ContractCatalogID,
		Command:    "updateMetadata",
		Args: byzcoin.Arguments{
			{
				Name: "metadataJSON", Value: []byte(metadataJSON),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return xerrors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return xerrors.New("failed to sign transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Dataset updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// CatalogGet checks the proof and prints the content of the catalog contract.
func CatalogGet(c *cli.Context) error {

	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	_, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	instID := c.String("instid")
	if instID == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instID)
	if err != nil {
		return errors.New("failed to decode the instID string")
	}

	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return errors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return errors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return errors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return errors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID, &catalogData)
	if err != nil {
		return errors.New("couldn't get a catalog instance: " + err.Error())
	}

	if c.Bool("export") {
		_, buf, _, _, err := proof.KeyValue()
		if err != nil {
			return errors.New("failed to get value from proof: " + err.Error())
		}
		reader := bytes.NewReader(buf)
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return errors.New("failed to copy to stdout: " + err.Error())
		}
		return nil
	}

	log.Infof("%s", catalogData)

	return nil
}

// CatalogGetDatasets checks the proof and prints the datasets of an owner.
func CatalogGetDatasets(c *cli.Context) error {

	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	_, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	instID := c.String("instid")
	if instID == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instID)
	if err != nil {
		return errors.New("failed to decode the instID string")
	}

	identityStr := c.String("identityStr")
	if identityStr == "" {
		return errors.New("please provide the identityStr with --identityStr")
	}

	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return errors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return errors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return errors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return errors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID, &catalogData)
	if err != nil {
		return errors.New("couldn't get a catalog instance: " + err.Error())
	}

	if catalogData.Owners == nil {
		return errors.New("CatalogData.Owners is nil, nothing to show then")
	}

	owner := catalogData.GetOwner(identityStr)
	if owner == nil {
		return xerrors.Errorf("owner with identity '%s' not found", identityStr)
	}

	if owner.Datasets == nil {
		owner.Datasets = make([]*catalogc.Dataset, 1)
	}

	if c.Bool("toJson") {
		jsonStr, err := json.Marshal(owner.Datasets)
		if err != nil {
			return xerrors.Errorf("failed to convert the datasets to json: %v", err)
		}

		log.Infof("%s", jsonStr)
		return nil
	}

	for _, dataset := range owner.Datasets {
		log.Infof("%s", dataset)
	}

	return nil
}

// CatalogGetSingleDataset checks the proof and prints the dataset if found
// among all the owner's datasets. Otherwise outputs an error.
func CatalogGetSingleDataset(c *cli.Context) error {

	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	_, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	instID := c.String("instid")
	if instID == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instID)
	if err != nil {
		return errors.New("failed to decode the instID string")
	}

	calypsoWriteID := c.String("calypsoWriteID")
	if instID == "" {
		return errors.New("--calypsoWriteID flag is required")
	}

	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return errors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return errors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return errors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return errors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID, &catalogData)
	if err != nil {
		return errors.New("couldn't get a catalog instance: " + err.Error())
	}

	if catalogData.Owners == nil {
		return errors.New("CatalogData.Owners is nil, nothing to show then")
	}

	var dataset *catalogc.Dataset

	for _, owner := range catalogData.Owners {
		dataset = owner.GetDataset(calypsoWriteID)
		if dataset != nil {
			break
		}
	}

	if dataset == nil {
		return xerrors.Errorf("datasets with calypsoWriteID '%s' not found",
			calypsoWriteID)
	}

	if c.Bool("export") {
		datsetBuf, err := protobuf.Encode(dataset)
		if err != nil {
			return errors.New("failed to encode dataset: " + err.Error())
		}
		reader := bytes.NewReader(datsetBuf)
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return errors.New("failed to copy to stdout: " + err.Error())
		}
		return nil
	}

	if c.Bool("toJson") {
		jsonStr, err := json.Marshal(dataset)
		if err != nil {
			return xerrors.Errorf("failed to convert the dataset to json: %v", err)
		}

		log.Infof("%s", jsonStr)
		return nil
	}

	log.Infof("%s", dataset)

	return nil
}

// CatalogGetMetadata checks the proof and prints the metadata.
func CatalogGetMetadata(c *cli.Context) error {

	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	_, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	instID := c.String("instid")
	if instID == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instID)
	if err != nil {
		return errors.New("failed to decode the instID string")
	}

	pr, err := cl.GetProofFromLatest(instIDBuf)
	if err != nil {
		return errors.New("couldn't get proof: " + err.Error())
	}
	proof := pr.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	if err != nil {
		return errors.New("error while checking if proof exist: " + err.Error())
	}
	if !exist {
		return errors.New("proof not found")
	}

	match := proof.InclusionProof.Match(instIDBuf)
	if !match {
		return errors.New("proof does not match")
	}

	var catalogData catalogc.CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, catalogc.ContractCatalogID, &catalogData)
	if err != nil {
		return errors.New("couldn't get a catalog instance: " + err.Error())
	}

	if catalogData.Metadata == nil {
		return errors.New("CatalogData.Metadata is nil, nothing to show then")
	}

	if c.Bool("export") {
		metadataBuf, err := protobuf.Encode(catalogData.Metadata)
		if err != nil {
			return errors.New("failed to encode metadata: " + err.Error())
		}
		reader := bytes.NewReader(metadataBuf)
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return errors.New("failed to copy to stdout: " + err.Error())
		}
		return nil
	}

	if c.Bool("toJson") {
		jsonStr, err := json.Marshal(catalogData.Metadata)
		if err != nil {
			return xerrors.Errorf("failed to convert the metadata to json: %v", err)
		}

		log.Infof("%s", jsonStr)
		return nil
	}

	log.Infof("%s", catalogData.Metadata)

	return nil
}
