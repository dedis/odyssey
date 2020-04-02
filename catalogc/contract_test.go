package catalogc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
)

func TestCatalogScenario(t *testing.T) {
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	counter := uint64(0)

	signer := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{"spawn:odysseycatalog", "invoke:odysseycatalog.addOwner",
			"invoke:odysseycatalog.updateOwner",
			"invoke:odysseycatalog.addDataset",
			"invoke:odysseycatalog.updateDataset",
			"invoke:odysseycatalog.deleteOwner",
			"invoke:odysseycatalog.updateMetadata",
			"invoke:odysseycatalog.deleteDataset",
			"invoke:odysseycatalog.archiveDataset"}, signer.Identity())
	require.Nil(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second

	cl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.Nil(t, err)

	// ------------------------------------------------------------------------
	// Spawn
	counter++
	ctx, err := cl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: ContractCatalogID,
			Args:       []byzcoin.Argument{},
		},
		SignerCounter: []uint64{counter},
	})
	require.NoError(t, err)
	require.Nil(t, ctx.FillSignersAndSignWith(signer))

	instID := ctx.Instructions[0].DeriveID("")
	instIDBuf := instID.Slice()

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)
	pr, err := cl.WaitProof(instID, 2*genesisMsg.BlockInterval, nil)
	require.Nil(t, err)

	require.True(t, pr.InclusionProof.Match(instIDBuf))
	_, _, _, err = pr.Get(instIDBuf)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err := cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof := prResp.Proof

	exist, err := proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match := proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	var catalogData CatalogData
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	require.Nil(t, catalogData.Owners)

	expected := `- Catalog:
-- Owners:
`
	require.Equal(t, expected, catalogData.String())

	// ------------------------------------------------------------------------
	// Add owner

	firstnameStr := "John"
	lastnameStr := "Doe"
	identityStr := "ed25519:aef123"

	invoke := byzcoin.Invoke{
		ContractID: ContractCatalogID,
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
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add owner with the same identity
	// (this should fail)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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
	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Error(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners := []*Owner{
		&Owner{
			Firstname:   firstnameStr,
			Lastname:    lastnameStr,
			IdentityStr: "ed25519:aef123",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: John
----- Lastname: Doe
----- IdentityStr: ed25519:aef123
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update the owner

	firstnameStr = "Frank"
	lastnameStr = "Underwood"
	identityStr = "ed25519:aef123"
	newIdentityStr := "darc:aef123"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update an unexisting owner
	// (this should fail)

	firstnameStr = "John"
	lastnameStr = "Doe"
	identityStr = "darc:deadbeef"
	newIdentityStr = "ed25519:aef123"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Error(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add a dataset to the owner

	identityStr = "darc:aef123"
	calypsoWriteID := "abcdef1234"
	dataset := Dataset{
		Title:       "title",
		Description: "description",
		CloudURL:    "example.com",
		SHA2:        "abcd",
	}

	datasetBuf, err := protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add a dataset with a previously used CalypsoWriteID
	// (this should fail)

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:       "title2",
		Description: "description2",
		CloudURL:    "example.com/2",
		SHA2:        "abcdef",
	}

	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Error(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:          "title",
		Description:    "description",
		CloudURL:       "example.com",
		SHA2:           "abcd",
		IdentityStr:    identityStr,
		CalypsoWriteID: calypsoWriteID,
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
			Datasets: []*Dataset{
				&dataset,
			},
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: title
-------- Description: description
-------- CloudURL: example.com
-------- SHA2: abcd
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update a dataset

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:       "title2",
		Description: "description2",
		CloudURL:    "example.com/2",
		SHA2:        "abcdef",
	}

	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "updateDataset",
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

	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Archive a dataset

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:       "title2",
		Description: "description2",
		CloudURL:    "example.com/2",
		SHA2:        "abcdef",
	}

	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update a non-existing dataset
	// (should fail)

	identityStr = "darc:aef123"
	calypsoWriteID = "aaaaaaaa"
	dataset = Dataset{
		Title:       "title2",
		Description: "description2",
		CloudURL:    "example.com/2",
		SHA2:        "xxxx",
	}

	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "updateDataset",
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Error(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:          "title2",
		Description:    "description2",
		CloudURL:       "example.com/2",
		SHA2:           "abcdef",
		IdentityStr:    identityStr,
		CalypsoWriteID: calypsoWriteID,
		IsArchived:     true,
		Metadata:       &Metadata{},
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
			Datasets: []*Dataset{
				&dataset,
			},
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: title2
-------- Description: description2
-------- CloudURL: example.com/2
-------- SHA2: abcdef
-------- IdentityStr: darc:aef123
-------- IsArchived: true
-------- Metadata:
--------- Metadata:
---------- AttributesGroups:
---------- DelegatedEnforcement:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update a second time the dataset

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:       "New title",
		Description: "new description",
		CloudURL:    "10.10.10.10/new.aes",
		SHA2:        "12345678abcd",
	}

	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "updateDataset",
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

	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:aef123"
	calypsoWriteID = "abcdef1234"
	dataset = Dataset{
		Title:          "New title",
		Description:    "new description",
		CloudURL:       "10.10.10.10/new.aes",
		SHA2:           "12345678abcd",
		IdentityStr:    identityStr,
		CalypsoWriteID: calypsoWriteID,
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
			Datasets: []*Dataset{
				&dataset,
			},
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: New title
-------- Description: new description
-------- CloudURL: 10.10.10.10/new.aes
-------- SHA2: 12345678abcd
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add a second dataset to the owner

	identityStr = "darc:aef123"
	calypsoWriteID = "11111111"
	dataset = Dataset{
		Title:       "other title",
		Description: "other description",
		CloudURL:    "other link",
		SHA2:        "other hash",
	}

	datasetBuf = []byte{}
	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:aef123"
	calypsoWriteID = "11111111"
	dataset = Dataset{
		Title:          "New title",
		Description:    "new description",
		CloudURL:       "10.10.10.10/new.aes",
		SHA2:           "12345678abcd",
		IdentityStr:    identityStr,
		CalypsoWriteID: "abcdef1234",
	}
	dataset2 := Dataset{
		Title:          "other title",
		Description:    "other description",
		CloudURL:       "other link",
		SHA2:           "other hash",
		IdentityStr:    identityStr,
		CalypsoWriteID: calypsoWriteID,
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
			Datasets: []*Dataset{
				&dataset,
				&dataset2,
			},
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: New title
-------- Description: new description
-------- CloudURL: 10.10.10.10/new.aes
-------- SHA2: 12345678abcd
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
------ Datasets[1]:
------- Dataset:
-------- CalypsoWriteID: 11111111
-------- Title: other title
-------- Description: other description
-------- CloudURL: other link
-------- SHA2: other hash
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add an existing dataset to the owner
	// (this should fail)

	identityStr = "darc:aef123"
	calypsoWriteID = "11111111"
	dataset = Dataset{
		Title:       "other title",
		Description: "other description",
		CloudURL:    "other link",
		SHA2:        "other hash",
	}

	datasetBuf = []byte{}
	datasetBuf, err = protobuf.Encode(&dataset)
	require.Nil(t, err)

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Error(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Add second owner

	firstnameStr = "Bob"
	lastnameStr = "Dylan"
	identityStr = "ed25519:aabbccdd"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:aef123"
	calypsoWriteID = "11111111"
	dataset = Dataset{
		Title:          "New title",
		Description:    "new description",
		CloudURL:       "10.10.10.10/new.aes",
		SHA2:           "12345678abcd",
		IdentityStr:    identityStr,
		CalypsoWriteID: "abcdef1234",
	}
	dataset2 = Dataset{
		Title:          "other title",
		Description:    "other description",
		CloudURL:       "other link",
		SHA2:           "other hash",
		IdentityStr:    identityStr,
		CalypsoWriteID: calypsoWriteID,
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:aef123",
			Datasets: []*Dataset{
				&dataset,
				&dataset2,
			},
		},
		&Owner{
			Firstname:   "Bob",
			Lastname:    "Dylan",
			IdentityStr: "ed25519:aabbccdd",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:aef123
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: New title
-------- Description: new description
-------- CloudURL: 10.10.10.10/new.aes
-------- SHA2: 12345678abcd
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
------ Datasets[1]:
------- Dataset:
-------- CalypsoWriteID: 11111111
-------- Title: other title
-------- Description: other description
-------- CloudURL: other link
-------- SHA2: other hash
-------- IdentityStr: darc:aef123
-------- IsArchived: false
-------- Metadata:
--- Owners[1]:
---- Owner:
----- Firstname: Bob
----- Lastname: Dylan
----- IdentityStr: ed25519:aabbccdd
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update the first identityStr's owner

	identityStr = "darc:aef123"
	newIdentityStr = "darc:bbbbbb"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "updateOwner",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
			{
				Name: "newIdentityStr", Value: []byte(newIdentityStr),
			},
		},
	}

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:bbbbbb"
	dataset = Dataset{
		Title:          "New title",
		Description:    "new description",
		CloudURL:       "10.10.10.10/new.aes",
		SHA2:           "12345678abcd",
		IdentityStr:    identityStr,
		CalypsoWriteID: "abcdef1234",
	}
	dataset2 = Dataset{
		Title:          "other title",
		Description:    "other description",
		CloudURL:       "other link",
		SHA2:           "other hash",
		IdentityStr:    identityStr,
		CalypsoWriteID: "11111111",
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:bbbbbb",
			Datasets: []*Dataset{
				&dataset,
				&dataset2,
			},
		},
		&Owner{
			Firstname:   "Bob",
			Lastname:    "Dylan",
			IdentityStr: "ed25519:aabbccdd",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:bbbbbb
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: abcdef1234
-------- Title: New title
-------- Description: new description
-------- CloudURL: 10.10.10.10/new.aes
-------- SHA2: 12345678abcd
-------- IdentityStr: darc:bbbbbb
-------- IsArchived: false
-------- Metadata:
------ Datasets[1]:
------- Dataset:
-------- CalypsoWriteID: 11111111
-------- Title: other title
-------- Description: other description
-------- CloudURL: other link
-------- SHA2: other hash
-------- IdentityStr: darc:bbbbbb
-------- IsArchived: false
-------- Metadata:
--- Owners[1]:
---- Owner:
----- Firstname: Bob
----- Lastname: Dylan
----- IdentityStr: ed25519:aabbccdd
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Delete the first dataset of the first owner

	identityStr = "darc:bbbbbb"
	calypsoWriteID = "abcdef1234"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
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

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	identityStr = "darc:bbbbbb"
	dataset2 = Dataset{
		Title:          "other title",
		Description:    "other description",
		CloudURL:       "other link",
		SHA2:           "other hash",
		IdentityStr:    identityStr,
		CalypsoWriteID: "11111111",
	}

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Frank",
			Lastname:    "Underwood",
			IdentityStr: "darc:bbbbbb",
			Datasets: []*Dataset{
				&dataset2,
			},
		},
		&Owner{
			Firstname:   "Bob",
			Lastname:    "Dylan",
			IdentityStr: "ed25519:aabbccdd",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Frank
----- Lastname: Underwood
----- IdentityStr: darc:bbbbbb
----- Datasets:
------ Datasets[0]:
------- Dataset:
-------- CalypsoWriteID: 11111111
-------- Title: other title
-------- Description: other description
-------- CloudURL: other link
-------- SHA2: other hash
-------- IdentityStr: darc:bbbbbb
-------- IsArchived: false
-------- Metadata:
--- Owners[1]:
---- Owner:
----- Firstname: Bob
----- Lastname: Dylan
----- IdentityStr: ed25519:aabbccdd
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Delete the first owner

	identityStr = "darc:bbbbbb"

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "deleteOwner",
		Args: byzcoin.Arguments{
			{
				Name: "identityStr", Value: []byte(identityStr),
			},
		},
	}

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Bob",
			Lastname:    "Dylan",
			IdentityStr: "ed25519:aabbccdd",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Bob
----- Lastname: Dylan
----- IdentityStr: ed25519:aabbccdd
----- Datasets:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Update the metadata

	metadataJSON := `{
		"attributesGroups": [{
				"title": "Use",
				"description": "How this dataset can be used",
				"attributes": [{
						"id": "use_restricted",
						"description": "The use of this dataset is restricted",
						"type": "checkbox",
						"attributes": [{
							"id": "use_restricted_description",
							"description": "Please describe the restriction",
							"type": "text",
							"value": "This is my custom restriction",
							"attributes": []
						}]
					},
					{
						"id": "use_predefined_purposes",
						"description": "Can be used for the following predefined-purposess",
						"type": "checkbox",
						"attributes": [{
							"id": "use_predefined_purposes_legal",
							"description": "To meet legal or regulatory requirements",
							"type": "checkbox",
							"attributes": []
						}, {
							"id": "use_predefined_purposes_analytics_counterparty",
							"description": "To provide analytics to the counterparty of the contract from which data is sourced/for the counterparty's benefit",
							"type": "checkbox",
							"attributes": []
						}]
					}
				]
			},
			{
				"title": "Access",
				"description": "Tell us who can access the data",
				"attributes": [{
					"id": "access_unrestricted",
					"description": "No restriction",
					"type": "radio",
					"name": "access",
					"attributes": []
				}, {
					"id": "access_defined_group",
					"description": "SR defined group",
					"type": "radio",
					"name": "access",
					"attributes": [{
						"id": "access_defined_group_description",
						"description": "Please specify the group",
						"type": "text"
					}]
				}]
			}
		]
	}`

	invoke = byzcoin.Invoke{
		ContractID: ContractCatalogID,
		Command:    "updateMetadata",
		Args: byzcoin.Arguments{
			{
				Name: "metadataJSON", Value: []byte(metadataJSON),
			},
		},
	}

	counter++
	ctx, err = cl.CreateTransaction(byzcoin.Instruction{
		InstanceID:    byzcoin.NewInstanceID(instIDBuf),
		Invoke:        &invoke,
		SignerCounter: []uint64{counter},
	})
	require.Nil(t, nil)

	err = ctx.FillSignersAndSignWith(signer)
	require.Nil(t, err)

	_, err = cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)

	local.WaitDone(genesisMsg.BlockInterval)

	// ------------------------------------------------------------------------
	// Get

	prResp, err = cl.GetProofFromLatest(instIDBuf)
	require.Nil(t, err)

	proof = prResp.Proof

	exist, err = proof.InclusionProof.Exists(instIDBuf)
	require.Nil(t, err)
	require.True(t, exist)

	match = proof.InclusionProof.Match(instIDBuf)
	require.True(t, match)

	catalogData = CatalogData{}
	err = proof.VerifyAndDecode(cothority.Suite, ContractCatalogID, &catalogData)
	require.Nil(t, err)

	owners = []*Owner{
		&Owner{
			Firstname:   "Bob",
			Lastname:    "Dylan",
			IdentityStr: "ed25519:aabbccdd",
		},
	}

	require.Equal(t, owners, catalogData.Owners)

	expected = `- Catalog:
-- Owners:
--- Owners[0]:
---- Owner:
----- Firstname: Bob
----- Lastname: Dylan
----- IdentityStr: ed25519:aabbccdd
----- Datasets:
-- Metadata:
--- AttributesGroups:
---- AttributesGroups[0]:
----- AttributesGroup:
------ Title: Use
------ Description: How this dataset can be used
------ ConsumerDescription: 
------ Attributes:
------- Attributes[0]:
-------- Attribute:
--------- ID: use_restricted
--------- Description: The use of this dataset is restricted
--------- Type: checkbox
--------- RuleType: 
--------- Name: 
--------- Value: 
--------- DelegatedEnforcement: false
--------- Attributes:
---------- Attributes[0]:
----------- Attribute:
------------ ID: use_restricted_description
------------ Description: Please describe the restriction
------------ Type: text
------------ RuleType: 
------------ Name: 
------------ Value: This is my custom restriction
------------ DelegatedEnforcement: false
------------ Attributes:
------- Attributes[1]:
-------- Attribute:
--------- ID: use_predefined_purposes
--------- Description: Can be used for the following predefined-purposess
--------- Type: checkbox
--------- RuleType: 
--------- Name: 
--------- Value: 
--------- DelegatedEnforcement: false
--------- Attributes:
---------- Attributes[0]:
----------- Attribute:
------------ ID: use_predefined_purposes_legal
------------ Description: To meet legal or regulatory requirements
------------ Type: checkbox
------------ RuleType: 
------------ Name: 
------------ Value: 
------------ DelegatedEnforcement: false
------------ Attributes:
---------- Attributes[1]:
----------- Attribute:
------------ ID: use_predefined_purposes_analytics_counterparty
------------ Description: To provide analytics to the counterparty of the contract from which data is sourced/for the counterparty's benefit
------------ Type: checkbox
------------ RuleType: 
------------ Name: 
------------ Value: 
------------ DelegatedEnforcement: false
------------ Attributes:
---- AttributesGroups[1]:
----- AttributesGroup:
------ Title: Access
------ Description: Tell us who can access the data
------ ConsumerDescription: 
------ Attributes:
------- Attributes[0]:
-------- Attribute:
--------- ID: access_unrestricted
--------- Description: No restriction
--------- Type: radio
--------- RuleType: 
--------- Name: access
--------- Value: 
--------- DelegatedEnforcement: false
--------- Attributes:
------- Attributes[1]:
-------- Attribute:
--------- ID: access_defined_group
--------- Description: SR defined group
--------- Type: radio
--------- RuleType: 
--------- Name: access
--------- Value: 
--------- DelegatedEnforcement: false
--------- Attributes:
---------- Attributes[0]:
----------- Attribute:
------------ ID: access_defined_group_description
------------ Description: Please specify the group
------------ Type: text
------------ RuleType: 
------------ Name: 
------------ Value: 
------------ DelegatedEnforcement: false
------------ Attributes:
--- DelegatedEnforcement:
`
	require.Equal(t, expected, catalogData.String())
	local.WaitDone(genesisMsg.BlockInterval)
}
