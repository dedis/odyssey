package clicontracts

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dedis/odyssey/projectc"
	"github.com/urfave/cli"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// ProjectSpawn spawns a new instance of a project contract
func ProjectSpawn(c *cli.Context) error {
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

	instIDstr := c.String("instids")
	if instIDstr == "" {
		return errors.New("please provide the instance IDs with --instids")
	}

	r, err := regexp.Compile("^[0-9a-f]{64}(,[0-9a-f]{64})*$")
	if err != nil {
		return errors.New("failed to build regex: " + err.Error())
	}
	instIDstr = strings.Trim(instIDstr, " \n\r")
	ok := r.MatchString(instIDstr)
	if !ok {
		return errors.New("Got unexpected 'instids': " + instIDstr)
	}

	pubKey := c.String("pubKey")

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	ctx := byzcoin.NewClientTransaction(byzcoin.CurrentVersion, byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(d.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: projectc.ContractProjectID,
			Args: byzcoin.Arguments{
				{
					Name: "datasetIDs", Value: []byte(instIDstr),
				},
				{
					Name: "accessPubKey", Value: []byte(pubKey),
				},
			},
		},
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})

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

	result := projectc.ProjectData{}
	err = protobuf.Decode(resultBuf, &result)
	if err != nil {
		return errors.New("couldn't decode the result: " + err.Error())
	}

	iidStr := hex.EncodeToString(instID.Slice())
	log.Infof("Spawned a new project instance. "+
		"Its instance id is:\n%s", iidStr)

	return lib.WaitPropagation(c, cl)
}

// ProjectdInvokeUpdate updates the project contract
func ProjectdInvokeUpdate(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	instIDstr := c.String("instid")
	if instIDstr == "" {
		return errors.New("--instid flag is required")
	}
	instIDBuf, err := hex.DecodeString(instIDstr)
	if err != nil {
		return errors.New("failed to decode the instid string: " + err.Error())
	}

	// We get the project data to update only what is given in argument
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

	var projectData projectc.ProjectData
	err = proof.VerifyAndDecode(cothority.Suite, projectc.ContractProjectID, &projectData)
	if err != nil {
		return errors.New("couldn't get a project instance: " + err.Error())
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

	instIDstr = c.String("instids")
	if instIDstr != "" {
		r, err := regexp.Compile("^[0-9a-f]{64}(,[0-9a-f]{64})*$")
		if err != nil {
			return errors.New("failed to build regex: " + err.Error())
		}
		instIDstr = strings.Trim(instIDstr, " \n\r")
		ok := r.MatchString(instIDstr)
		if !ok {
			return errors.New("Got unexpected 'instids': " + instIDstr)
		}

		instIDlist := strings.Split(instIDstr, ",")

		datasets := make([]byzcoin.InstanceID, len(instIDlist))
		for i, instID := range instIDlist {
			decodedInstID, err := hex.DecodeString(instID)
			if err != nil {
				return errors.New("failed to decode instance id: " + err.Error())
			}
			datasets[i] = byzcoin.NewInstanceID(decodedInstID)
		}
		projectData.Datasets = datasets
	}

	projectDataBuf, err := protobuf.Encode(&projectData)
	if err != nil {
		return errors.New("failed to encode projectData: " + err.Error())
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: projectc.ContractProjectID,
		Command:    "update",
		Args: []byzcoin.Argument{
			{
				Name:  "projectData",
				Value: projectDataBuf,
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID([]byte(instIDBuf)),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	if lib.FindRecursivefBool("export", c) {
		return lib.ExportTransaction(ctx)
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction and wait: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Value contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectdInvokeUpdateStatus update the status of the project. Returns an error
// if it doesn't match any status specified in the contract.
func ProjectdInvokeUpdateStatus(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
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

	status := c.String("status")
	if status == "" {
		return errors.New("please provide an enclave URL with --status")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: projectc.ContractProjectID,
		Command:    "updateStatus",
		Args: []byzcoin.Argument{
			{
				Name:  "status",
				Value: []byte(status),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID([]byte(instIDBuf)),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction and wait: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Value contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectInvokeUpdateMetadata upadte the metadata
func ProjectInvokeUpdateMetadata(c *cli.Context) error {
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
		ContractID: projectc.ContractProjectID,
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
	fmt.Printf("Project updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectdInvokeSetURL set the EnclaveURL attributes
func ProjectdInvokeSetURL(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
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

	enclaveURL := c.String("enclaveURL")
	if enclaveURL == "" {
		return errors.New("please provide an enclave URL with --enclaveURL")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: projectc.ContractProjectID,
		Command:    "setURL",
		Args: []byzcoin.Argument{
			{
				Name:  "enclaveURL",
				Value: []byte(enclaveURL),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID([]byte(instIDBuf)),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction and wait: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Value contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectdInvokeSetAccessPubKey set the EnclaveURL attributes
func ProjectdInvokeSetAccessPubKey(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
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

	pubKey := c.String("pubKey")
	if pubKey == "" {
		return errors.New("please provide an enclave URL with --pubKey")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: projectc.ContractProjectID,
		Command:    "setAccessPubKey",
		Args: []byzcoin.Argument{
			{
				Name:  "pubKey",
				Value: []byte(pubKey),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID([]byte(instIDBuf)),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction and wait: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Value contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectdInvokeSetEnclavePubKey set the EnclaveURL attributes
func ProjectdInvokeSetEnclavePubKey(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return errors.New("--bc flag is required")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
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

	pubKey := c.String("pubKey")
	if pubKey == "" {
		return errors.New("please provide an enclave URL with --pubKey")
	}

	counters, err := cl.GetSignerCounters(signer.Identity().String())

	invoke := byzcoin.Invoke{
		ContractID: projectc.ContractProjectID,
		Command:    "setEnclavePubKey",
		Args: []byzcoin.Argument{
			{
				Name:  "pubKey",
				Value: []byte(pubKey),
			},
		},
	}

	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID([]byte(instIDBuf)),
		Invoke:        &invoke,
		SignerCounter: []uint64{counters.Counters[0] + 1},
	})
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	err = ctx.FillSignersAndSignWith(*signer)
	if err != nil {
		return errors.New("failed to create transaction: " + err.Error())
	}

	_, err = cl.AddTransactionAndWait(ctx, 10)
	if err != nil {
		return errors.New("failed to add transaction and wait: " + err.Error())
	}

	newInstID := ctx.Instructions[0].DeriveID("").Slice()
	fmt.Printf("Value contract updated! (instance ID is %x)\n", newInstID)

	return lib.WaitPropagation(c, cl)
}

// ProjectGet checks the proof and prints the content of the Write contract.
func ProjectGet(c *cli.Context) error {

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

	var projectData projectc.ProjectData
	err = proof.VerifyAndDecode(cothority.Suite, projectc.ContractProjectID, &projectData)
	if err != nil {
		return errors.New("couldn't get a project instance: " + err.Error())
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

	log.Infof("%s", projectData)

	return nil
}
