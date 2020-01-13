package main

import (
	"github.com/dedis/odyssey/projectc/pcadmin/clicontracts"
	"github.com/urfave/cli"
)

var cmds = cli.Commands{
	{
		Name:  "contract",
		Usage: "Provides cli interface for project contract",
		Subcommands: cli.Commands{
			{
				Name:  "project",
				Usage: "handles project contract",
				Subcommands: cli.Commands{
					{
						Name:   "spawn",
						Usage:  "spawn a project contract.",
						Action: clicontracts.ProjectSpawn,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:   "bc",
								EnvVar: "BC",
								Usage:  "the ByzCoin config to use (required)",
							},
							cli.StringFlag{
								Name:  "darc",
								Usage: "DARC with the right to create a Write instance (default is the admin DARC)",
							},
							cli.StringFlag{
								Name:  "sign, s",
								Usage: "public key of the signing entity (default is the admin)",
							},
							cli.StringFlag{
								Name:  "instids, is",
								Usage: "the write instance ids, separeted by comas",
							},
							cli.StringFlag{
								Name:  "pubKey, pk",
								Usage: "an RSA public key string of type 'ssh-rsa XXX...' (optional)",
							},
							cli.StringFlag{
								Name:  "metadataJson, mJson",
								Usage: "the metadata in JSON format (required)",
							},
						},
					},
					{
						Name:  "invoke",
						Usage: "invoke on a project contract.",
						Subcommands: cli.Commands{
							{
								Name:   "update",
								Usage:  "update a project contract with the given project data fields",
								Action: clicontracts.ProjectdInvokeUpdate,
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
										Name:  "darc",
										Usage: "DARC with the right to invoke.update a project contract (default is the admin DARC)",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "instids, is",
										Usage: "the write instance ids, separeted by comas (fills the 'Datasets' field)",
									},
								},
							},
							{
								Name:   "updateStatus",
								Usage:  "update the status attribute of the project",
								Action: clicontracts.ProjectdInvokeUpdateStatus,
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
										Name:  "darc",
										Usage: "DARC with the right to invoke.updateStatus a project contract (default is the admin DARC)",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "status",
										Usage: "the status in its string form. Return an error if it doesn't match any status described in the contract",
									},
								},
							},
							{
								Name:   "updateMetadata",
								Usage:  "set the metadata to the given JSON representation",
								Action: clicontracts.ProjectInvokeUpdateMetadata,
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
							{
								Name:   "setURL",
								Usage:  "sets the EnclaveURL attribute of the project",
								Action: clicontracts.ProjectdInvokeSetURL,
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
										Name:  "darc",
										Usage: "DARC with the right to invoke.setURL a project contract (default is the admin DARC)",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "enclaveURL, eu",
										Usage: "the enclave URL that can be used to ssh to it",
									},
								},
							},
							{
								Name:   "setAccessPubKey",
								Usage:  "sets the accessPubKey attribute of the project. This key will have the final access to then enclave",
								Action: clicontracts.ProjectdInvokeSetAccessPubKey,
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
										Name:  "darc",
										Usage: "DARC with the right to invoke.setAccessPubKey a project contract (default is the admin DARC)",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "pubKey, pk",
										Usage: "the public key that can be used to ssh to it",
									},
								},
							},
							{
								Name:   "setEnclavePubKey",
								Usage:  "sets the enclavePubKey attribute of the project. This key will be able to update the instance of this contract",
								Action: clicontracts.ProjectdInvokeSetEnclavePubKey,
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
										Name:  "darc",
										Usage: "DARC with the right to invoke.setEnclavePubKey a project contract (default is the admin DARC)",
									},
									cli.StringFlag{
										Name:  "sign, s",
										Usage: "public key of the signing entity (default is the admin public key)",
									},
									cli.StringFlag{
										Name:  "pubKey, pk",
										Usage: "the public key that can be used to update the instance of this contract",
									},
								},
							},
						},
					},
					{
						Name:   "get",
						Usage:  "if the proof matches, prints the content of the given project instance ID",
						Action: clicontracts.ProjectGet,
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
				},
			},
		},
	},
}
