package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/projectc"
	"github.com/urfave/cli"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
	"go.dedis.ch/onet/v3/cfgpath"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

var cliApp = cli.NewApp()

// getDataPath is a function pointer so that tests can hook and modify this.
var getDataPath = cfgpath.GetDataPath

var gitTag = "dev"

func init() {
	cliApp.Name = "catadmin"
	cliApp.Usage = "Handle the catalog contract service"
	cliApp.Version = gitTag
	cliApp.Commands = cmds // stored in "commands.go"
	cliApp.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
		cli.StringFlag{
			Name:   "config, c",
			EnvVar: "BC_CONFIG",
			// We use the bcadmin config folder because the bcadmin utiliy is
			// the prefered way to generate the config files. And this is where
			// bcadmin will put them.
			Value: getDataPath(lib.BcaName),
			Usage: "path to configuration-directory",
		},
		cli.BoolFlag{
			Name:   "wait, w",
			EnvVar: "BC_WAIT",
			Usage:  "wait for transaction available in all nodes",
		},
	}
	cliApp.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		lib.ConfigPath = c.String("config")
		return nil
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	err := cliApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// auditDataset prints all the access performed on the dataset
func auditDataset(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return xerrors.New("--bc flag is required")
	}

	instid := c.String("instid")
	if instid == "" {
		return xerrors.New("please provide a Calypso write instanceID with --instid")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	msg := &byzcoin.PaginateRequest{
		StartID:  cfg.ByzCoinID,
		PageSize: 1,
		NumPages: 100000,
		Backward: false,
	}
	ret := &byzcoin.PaginateResponse{}
	streamingCon, err := cl.Stream(cfg.Roster.RandomServerIdentity(), msg)
	if err != nil {
		return xerrors.Errorf("failed to call PaginateRequest: %v", err)
	}

	out := new(strings.Builder)
	occurences := 0
	nblocks := 0

	for ; nblocks < 100000; nblocks++ {
		err = streamingCon.ReadMessage(ret)
		if err != nil {
			return xerrors.Errorf("failed to read from stream: %v", err)
		}
		// This is normal when it reaches the end of the chain
		if ret.ErrorCode == 4 {
			break
		}
		if ret.ErrorCode != 0 {
			return xerrors.Errorf("Got a non zero error code: %d, %v", ret.ErrorCode, ret.ErrorText)

		}
		if len(ret.Blocks) == 0 {
			return xerrors.Errorf("Expected to have one block, but got: %v", ret.Blocks)
		}
		dataBody := &byzcoin.DataBody{}
		err := protobuf.Decode(ret.Blocks[0].Payload, dataBody)
		if err != nil {
			return xerrors.Errorf("failed to decode dataBody: %v", err)
		}
		if len(dataBody.TxResults) == 0 {
			continue
		}
		for _, txResult := range dataBody.TxResults {
			for _, instr := range txResult.ClientTransaction.Instructions {
				if instr.InstanceID.String() == instid {
					occurences++
					out.WriteString("<div class=\"occurence\">")
					fmt.Fprintf(out, "<p>Accepted? <b>%v</b><br>", txResult.Accepted)
					fmt.Fprintf(out, "BlockIndex: %d<p>", ret.Blocks[0].Index)
					if instr.GetType() == byzcoin.SpawnType {
						projectInstID := instr.Spawn.Args.Search("projectInstID")
						fmt.Fprintf(out, "<p>Project instance ID: <a href='/lifecycle?piid=%x'>%x</a></p>", projectInstID, projectInstID)
						resp, err := cl.GetProofFromLatest(projectInstID)
						if err != nil {
							return xerrors.Errorf("failed to get project instance: %v", err)
						}
						var projectInst projectc.ProjectData
						err = resp.Proof.VerifyAndDecode(cothority.Suite, projectc.ContractProjectID, &projectInst)
						if err != nil {
							return xerrors.Errorf("failed to decode project instance: %v", err)
						}
						out.WriteString("<details>")
						out.WriteString("<summary>See the project attributes</summary>")
						fmt.Fprintf(out, "<pre>%v</pre>", projectInst)
						out.WriteString("</details>")
						out.WriteString("</details>")
					}
					fmt.Fprintf(out, "<details><summary>See instruction</summary><pre>%v</pre></details>", instr)
					out.WriteString("</div>")
				}
			}
		}
	}

	prolog := fmt.Sprintf("<p>Checked <b>%d blocks</b> and found <b>%d requests</b>.</p>", nblocks, occurences)
	log.Info(prolog + out.String())

	return nil
}

// auditProject prints all the access performed on the dataset
func auditProject(c *cli.Context) error {
	bcArg := c.String("bc")
	if bcArg == "" {
		return xerrors.New("--bc flag is required")
	}

	instid := c.String("instid")
	if instid == "" {
		return xerrors.New("please provide a Calypso write instanceID with --instid")
	}

	cfg, cl, err := lib.LoadConfig(bcArg)
	if err != nil {
		return err
	}

	msg := &byzcoin.PaginateRequest{
		StartID:  cfg.ByzCoinID,
		PageSize: 1,
		NumPages: 100000,
		Backward: false,
	}
	ret := &byzcoin.PaginateResponse{}
	streamingCon, err := cl.Stream(cfg.Roster.RandomServerIdentity(), msg)
	if err != nil {
		return xerrors.Errorf("failed to call PaginateRequest: %v", err)
	}

	blocks := make([]*catalogc.AuditBlock, 0)
	occurences := 0
	nblocks := 0

	for ; nblocks < 100000; nblocks++ {
		err = streamingCon.ReadMessage(ret)
		if err != nil {
			return xerrors.Errorf("failed to read from stream: %v", err)
		}
		// This is normal when it reaches the end of the chain
		if ret.ErrorCode == 4 {
			break
		}
		if ret.ErrorCode != 0 {
			return xerrors.Errorf("Got a non zero error code: %d, %v", ret.ErrorCode, ret.ErrorText)

		}
		if len(ret.Blocks) == 0 {
			return xerrors.Errorf("Expected to have one block, but got: %v", ret.Blocks)
		}
		dataBody := &byzcoin.DataBody{}
		err := protobuf.Decode(ret.Blocks[0].Payload, dataBody)
		if err != nil {
			return xerrors.Errorf("failed to decode dataBody: %v", err)
		}
		if len(dataBody.TxResults) == 0 {
			continue
		}

		transactionsRes := make([]*catalogc.AuditTransaction, 0)
		for _, txResult := range dataBody.TxResults {
			instrRes := make([]*byzcoin.Instruction, 0)

			for _, instr := range txResult.ClientTransaction.Instructions {
				if instr.InstanceID.String() == instid || instr.DeriveID("").String() == instid {
					instrRes = append(instrRes, &instr)
					occurences++
				}
			}

			if len(instrRes) != 0 {
				auditTransaction := &catalogc.AuditTransaction{
					Accepted:     txResult.Accepted,
					Instructions: instrRes,
				}
				transactionsRes = append(transactionsRes, auditTransaction)
			}
		}

		if len(transactionsRes) != 0 {
			auditBlock := &catalogc.AuditBlock{
				BlockIndex:   ret.Blocks[0].Index,
				Transactions: transactionsRes,
			}
			blocks = append(blocks, auditBlock)
		}
	}

	if len(blocks) > 0 {
		blocks[0].DeltaPrevious = -1
		blocks[len(blocks)-1].DeltaNext = -1
		for i, block := range blocks[1:] {
			block.DeltaPrevious = block.BlockIndex - blocks[i].BlockIndex - 1
		}
		for i, block := range blocks[:len(blocks)-1] {
			block.DeltaNext = blocks[i+1].BlockIndex - block.BlockIndex - 1
		}
	}

	result := catalogc.AuditData{
		BlocksChecked: nblocks,
		OccFound:      occurences,
		Blocks:        blocks,
	}

	resultBuf, err := protobuf.Encode(&result)
	if err != nil {
		return xerrors.Errorf("failed to encode audit result: %v", err)
	}

	reader := bytes.NewReader(resultBuf)
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return xerrors.Errorf("failed to export to stdout: %v", err)
	}

	return nil
}
