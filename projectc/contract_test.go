package projectc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
)

func TestProjectScenario(t *testing.T) {
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	counter := uint64(0)

	signer := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{"spawn:odysseyproject", "invoke:odysseyproject.update",
			"invoke:odysseyproject.updateStatus",
			"invoke:odysseyproject.updateMetadata",
			"invoke:odysseyproject.setURL",
			"invoke:odysseyproject.setAccessPubKey",
			"invoke:odysseyproject.setEnclavePubKey"}, signer.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second

	cl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// Spawn
	instID1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	instID2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	pubKey := "TEST_PUBKEY"

	counter++
	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: ContractProjectID,
			Args: []byzcoin.Argument{
				{Name: "datasetIDs", Value: []byte(instID1 + "," + instID2)},
				{Name: "accessPubKey", Value: []byte(pubKey)},
			},
		},
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)
	require.Nil(t, ctx.FillSignersAndSignWith(signer))

	instID := ctx.Instructions[0].DeriveID("")
	instIDBuf := instID.Slice()

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)
	pr, err := cl.WaitProof(instID, 2*genesisMsg.BlockInterval, nil)
	require.NoError(t, err)

	require.True(t, pr.InclusionProof.Match(instIDBuf))
	_, _, _, err = pr.Get(instIDBuf)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err := cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof := prResp.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match := proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	var projectData ProjectData
	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, pubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, empty, projectData.Status)

	// ------------------------------------------------------------------------
	// Update

	projectData.Status = initialized

	prjectDataBuf, err := protobuf.Encode(&projectData)
	require.NoError(t, err)

	invoke := byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "update",
		Args: byzcoin.Arguments{
			{
				Name: "projectData", Value: prjectDataBuf,
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, pubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, initialized, projectData.Status)

	// ------------------------------------------------------------------------
	// Update status

	newStatus := preparing

	invoke = byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "updateStatus",
		Args: byzcoin.Arguments{
			{
				Name: "status", Value: []byte(newStatus.String()),
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, pubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, newStatus, projectData.Status)

	// ------------------------------------------------------------------------
	// Update metadata

	metadata := catalogc.Metadata{}
	metadata.AttributesGroups = []*catalogc.AttributesGroup{
		{Title: "TEST_TITLE"},
	}
	metadataJSON, err := json.Marshal(&metadata)
	require.NoError(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "updateMetadata",
		Args: byzcoin.Arguments{
			{
				Name: "metadataJSON", Value: metadataJSON,
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, pubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, 1, len(projectData.Metadata.AttributesGroups))
	require.Equal(t, "TEST_TITLE", projectData.Metadata.AttributesGroups[0].Title)
	require.Equal(t, newStatus, projectData.Status)

	// ------------------------------------------------------------------------
	// SetURL

	url := "TEST_URL"

	invoke = byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "setURL",
		Args: byzcoin.Arguments{
			{
				Name: "enclaveURL", Value: []byte(url),
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, pubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, 1, len(projectData.Metadata.AttributesGroups))
	require.Equal(t, "TEST_TITLE", projectData.Metadata.AttributesGroups[0].Title)
	require.Equal(t, newStatus, projectData.Status)
	require.Equal(t, url, projectData.EnclaveURL)

	// ------------------------------------------------------------------------
	// SetAccessPubKey

	newAccessPubKey := "NEW_KEY"

	invoke = byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "setAccessPubKey",
		Args: byzcoin.Arguments{
			{
				Name: "pubKey", Value: []byte(newAccessPubKey),
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, newAccessPubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, 1, len(projectData.Metadata.AttributesGroups))
	require.Equal(t, "TEST_TITLE", projectData.Metadata.AttributesGroups[0].Title)
	require.Equal(t, newStatus, projectData.Status)
	require.Equal(t, url, projectData.EnclaveURL)

	// ------------------------------------------------------------------------
	// setEnclavePubKey

	enclaveKey := "ENCLAVE_KEY"

	invoke = byzcoin.Invoke{
		ContractID: ContractProjectID,
		Command:    "setEnclavePubKey",
		Args: byzcoin.Arguments{
			{
				Name: "pubKey", Value: []byte(enclaveKey),
			},
		},
	}
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)

	err = ctx.FillSignersAndSignWith(signer)
	require.NoError(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.NoError(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.NoError(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.NoError(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	err = proof.VerifyAndDecode(cothority.Suite, ContractProjectID, &projectData)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectData.Datasets))
	require.Equal(t, instID1, projectData.Datasets[0].String())
	require.Equal(t, instID2, projectData.Datasets[1].String())
	require.Equal(t, newAccessPubKey, projectData.AccessPubKey)
	require.NotNil(t, projectData.Metadata)
	require.Equal(t, 1, len(projectData.Metadata.AttributesGroups))
	require.Equal(t, "TEST_TITLE", projectData.Metadata.AttributesGroups[0].Title)
	require.Equal(t, newStatus, projectData.Status)
	require.Equal(t, url, projectData.EnclaveURL)
	require.Equal(t, enclaveKey, projectData.EnclavePubKey)
}
