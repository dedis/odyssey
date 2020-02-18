package main

import (
	"github.com/dedis/odyssey/catalogc/catadmin/clicontracts"
	"github.com/urfave/cli"
)

var cmds = cli.Commands{
	{
		Name:  "contract",
		Usage: "Provides cli interface for catalog contract",
		Subcommands: cli.Commands{
			{
				Name:  "catalog",
				Usage: "handles catalog contract",
				Subcommands: cli.Commands{
					{
						Name:   "spawn",
						Usage:  "spawn a catalog contract.",
						Action: clicontracts.CatalogSpawn,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "darc",
								Usage: "DARC with the right to create a Catalog instance (default is the admin DARC)",
							},
							cli.StringFlag{
								Name:  "sign, s",
								Usage: "public key of the signing entity (default is the admin)",
							},
						},
					},
					{
						Name:  "invoke",
						Usage: "invoke on a catalog contract.",
						Subcommands: cli.Commands{
							{
								Name:   "addOwner",
								Usage:  "add a new Owner to the catalog",
								Action: clicontracts.CatalogInvokeAddOwner,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "firstname, fn",
										Usage: "The firstname of the owner",
									},
									cli.StringFlag{
										Name:  "lastname, ln",
										Usage: "The lastname of the owner",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123'",
									},
								},
							},
							{
								Name:   "updateOwner",
								Usage:  "update the attributes of an existing Owner",
								Action: clicontracts.CatalogInvokeUpdateOwner,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
									cli.StringFlag{
										Name:  "firstname, fn",
										Usage: "The firstname of the owner (optional)",
									},
									cli.StringFlag{
										Name:  "lastname, ln",
										Usage: "The lastname of the owner (optional)",
									},
									cli.StringFlag{
										Name:  "newIdentityStr, nIdStr",
										Usage: "a new identity of the owner, like 'ed25519:aef123' (optional)",
									},
								},
							},
							{
								Name:   "deleteOwner",
								Usage:  "delete an existing Owner",
								Action: clicontracts.CatalogInvokeDeleteOwner,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
								},
							},
							{
								Name:   "addDataset",
								Usage:  "add a dataset to an Owner",
								Action: clicontracts.CatalogInvokeAddDataset,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
									cli.StringFlag{
										Name:  "calypsoWriteID",
										Usage: "calypsoWriteID of the dataset",
									},
									cli.StringFlag{
										Name:  "title",
										Usage: "title of the dataset",
									},
									cli.StringFlag{
										Name:  "description",
										Usage: "description of the dataset",
									},
									cli.StringFlag{
										Name:  "cloudURL",
										Usage: "cloudURL of the dataset",
									},
									cli.StringFlag{
										Name:  "sha2",
										Usage: "sha2 of the dataset",
									},
								},
							},
							{
								Name:   "updateDataset",
								Usage:  "update a dataset to an Owner",
								Action: clicontracts.CatalogInvokeUpdateDataset,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
									cli.StringFlag{
										Name:  "calypsoWriteID",
										Usage: "calypsoWriteID of the dataset",
									},
									cli.StringFlag{
										Name:  "newCalypsoWriteID",
										Usage: "newCalypsoWriteID of the dataset",
									},
									cli.StringFlag{
										Name:  "title",
										Usage: "title of the dataset",
									},
									cli.StringFlag{
										Name:  "description",
										Usage: "description of the dataset",
									},
									cli.StringFlag{
										Name:  "cloudURL",
										Usage: "cloudURL of the dataset",
									},
									cli.StringFlag{
										Name:  "sha2",
										Usage: "sha2 of the dataset",
									},
									cli.StringFlag{
										Name:  "metadataJSON, mJSON",
										Usage: "the JSON representation of the Metadata struct",
									},
								},
							},
							{
								Name:   "archiveDataset",
								Usage:  "delete a dataset from an Owner",
								Action: clicontracts.CatalogInvokeArchiveDataset,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
									cli.StringFlag{
										Name:  "calypsoWriteID",
										Usage: "calypsoWriteID of the dataset",
									},
								},
							},
							{
								Name:   "deleteDataset",
								Usage:  "delete a dataset from an Owner",
								Action: clicontracts.CatalogInvokeDeleteDataset,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "identityStr, idStr",
										Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
									},
									cli.StringFlag{
										Name:  "calypsoWriteID",
										Usage: "calypsoWriteID of the dataset",
									},
								},
							},
							{
								Name:   "updateMetadata",
								Usage:  "set the metadata to the given JSON representation",
								Action: clicontracts.CatalogInvokeUpdateMetadata,
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:   "bc",
										EnvVar: "BC",
										Usage:  "the ByzCoin config to use (required)",
									},
									cli.StringFlag{
										Name:  "instid, i",
										Usage: "the instance ID of the project contract",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "metadataJSON, mJSON",
										Usage: "the JSON representation of the Metadata struct (required)",
									},
								},
							},
						},
					},
					{
						Name:   "get",
						Usage:  "if the proof matches, prints the content of the given catalog instance ID",
						Action: clicontracts.CatalogGet,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "instid, i",
								Usage: "the instance id (required)",
							},
							cli.BoolFlag{
								Name:  "export, x",
								Usage: "export the write instance to STDOUT",
							},
						},
					},
					{
						Name:   "getDatasets",
						Usage:  "if the proof matches, prints the datasets of an Owner",
						Action: clicontracts.CatalogGetDatasets,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "instid, i",
								Usage: "the instance id (required)",
							},
							cli.BoolFlag{
								Name:  "toJson",
								Usage: "prints a json representation",
							},
							cli.StringFlag{
								Name:  "identityStr, idStr",
								Usage: "the identity of the owner, like 'ed25519:aef123' (required)",
							},
						},
					},
					{
						Name:   "getSingleDataset",
						Usage:  "if the proof matches, searches the dataset among all the owner's datasets",
						Action: clicontracts.CatalogGetSingleDataset,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "instid, i",
								Usage: "the instance id (required)",
							},
							cli.StringFlag{
								Name:  "calypsoWriteID, cwID",
								Usage: "the calypso write ID (required)",
							},
							cli.BoolFlag{
								Name:  "export, x",
								Usage: "sends the result to stdout",
							},
							cli.BoolFlag{
								Name:  "toJson",
								Usage: "prints a json representation, has no effect if --export is used",
							},
						},
					},
					{
						Name:   "getMetadata",
						Usage:  "if the proof matches, returns the Metadata",
						Action: clicontracts.CatalogGetMetadata,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "instid, i",
								Usage: "the instance id (required)",
							},
							cli.BoolFlag{
								Name:  "export, x",
								Usage: "sends the result to stdout",
							},
							cli.BoolFlag{
								Name:  "toJson",
								Usage: "prints a json representation, has no effect if --export is used",
							},
						},
					},
				},
			},
		},
	},
	{
		Name:  "audit",
		Usage: "Audit access on a dataset or display the lifecycle of a project",
		Subcommands: cli.Commands{
			{
				Name:   "dataset",
				Usage:  "display in HTML format all the access performed on the given Calypso write ID",
				Action: auditDataset,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bc",
						EnvVar: "BC",
						Usage:  "the ByzCoin config to use (required)",
					},
					cli.StringFlag{
						Name:  "instid, i",
						Usage: "The Calypso write ID",
					},
				},
			},
			{
				Name:   "project",
				Usage:  "display in HTML format the evolution of the project, ie. all the instructions concerned by the given instanceID",
				Action: auditProject,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bc",
						EnvVar: "BC",
						Usage:  "the ByzCoin config to use (required)",
					},
					cli.StringFlag{
						Name:  "instid, i",
						Usage: "The project instance ID",
					},
				},
			},
		},
	},
}
