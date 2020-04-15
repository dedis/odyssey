// This package provides a utility command line to encrypt and decrypt data with
// the "crypto/aes" library on AES-128 GCM.
// Install with "go install" and see help with "cryptutil -h"

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli"
)

var cliApp = cli.NewApp()

var gitTag = "dev"

func init() {
	cliApp.Name = "cryptutil"
	cliApp.Usage = "Encrypt and decrypt data with AES-128 GCM."
	cliApp.Version = gitTag
	cliApp.Commands = cli.Commands{
		cli.Command{
			Name:    "encrypt",
			Aliases: []string{"e"},
			Usage:   "Encrypts data and prints the hexadecimal representation of encrypted data. If using --export, outputs the raw encrypted data.",
			Action:  encrypt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "data, d",
					Usage: "The data to encrypt",
				},
				cli.StringFlag{
					Name:  "key, k",
					Usage: "the 128 bit key to use encoded as hexadecimal string (= 16 hex chars)",
				},
				cli.StringFlag{
					Name:  "initVal, iv",
					Usage: "The 96 bits initialization value encoded as hexadecimal string (= 12 hex chars)",
				},
				cli.StringFlag{
					Name:  "keyAndInitVal",
					Usage: "The key and initialization value as one 56 chars hex string (key || initVal). If used, the arguments --key and --initVal are not used.",
				},
				cli.BoolFlag{
					Name:  "readData, rd",
					Usage: "Read data from stdin. Cannot be used with --data",
				},
				cli.BoolFlag{
					Name:  "export, x",
					Usage: "do not print the encrypted data but sends it to stdout",
				},
			},
		},
		cli.Command{
			Name:    "decrypt",
			Aliases: []string{"d"},
			Usage:   "Decrypts data",
			Action:  decrypt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "data, d",
					Usage: "Encrypted data as hexadecimal string",
				},
				cli.StringFlag{
					Name:  "key, k",
					Usage: "the 128 bit key to use encoded in hexadecimal (32 hex chars)",
				},
				cli.StringFlag{
					Name:  "initVal, iv",
					Usage: "The 96 bits initialization value encoded in hexadecimal (24 hex char)",
				},
				cli.StringFlag{
					Name:  "keyAndInitVal",
					Usage: "The key and initialization value as one 56 chars hex string (key || initVal). If used, the arguments --key and --initVal are not used.",
				},
				cli.BoolFlag{
					Name:  "readData, rd",
					Usage: "Read raw encrypted data from stdin. Cannot be used with --data",
				},
				cli.BoolFlag{
					Name:  "export, x",
					Usage: "do not print the encrypted data but sends it to stdout",
				},
			},
		},
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	err := cliApp.Run(os.Args)
	if err != nil {
		log.Fatal("failed to run app: " + err.Error())
	}
}

func encrypt(c *cli.Context) error {

	var dataBuf []byte
	var err error

	if c.Bool("readData") {
		dataBuf, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errors.New("failed to read data from stdin")
		}
	} else {
		data := c.String("data")
		if data == "" {
			return errors.New("please provide data with --data")
		}
		dataBuf = []byte(data)
	}

	var key, iv string
	var keyBuf, ivBuf []byte

	keyAndInitVal := c.String("keyAndInitVal")
	if keyAndInitVal != "" {
		keyAndInitValBuf, err := hex.DecodeString(keyAndInitVal)
		if err != nil {
			return errors.New("failed to decode keyAndInitVal: " + err.Error())
		}
		// 16 bytes = 128 bits, 12 bytes = 96 bits
		if len(keyAndInitValBuf) != 16+12 {
			return fmt.Errorf("length of key + initVal must be 224 bits, not %d",
				len(keyAndInitValBuf)*8)
		}
		keyBuf = keyAndInitValBuf[:16]
		ivBuf = keyAndInitValBuf[16:]
	} else {
		key = c.String("key")
		if key == "" {
			return errors.New("please provide a key with --key")
		}

		iv = c.String("initVal")
		if iv == "" {
			return errors.New("please provice an initialization value with --initVal")
		}

		keyBuf, err = hex.DecodeString(key)
		if err != nil {
			return errors.New("failed to decode key as hexadecimal: " + err.Error())
		}
		if len(keyBuf) != 16 {
			return fmt.Errorf("length of key must be 128 bits, not %d",
				len(keyBuf)*8)
		}

		ivBuf, err = hex.DecodeString(iv)
		if err != nil {
			return errors.New("failed to decode initVal as hexadecimal: " + err.Error())
		}
		if len(ivBuf) != 12 {
			return fmt.Errorf("length of initialization value must be 96 bits, not %d",
				len(ivBuf)*8)
		}
	}

	block, err := aes.NewCipher(keyBuf)
	if err != nil {
		return errors.New("failed to create new cipher: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.New("failed to create new cipher: " + err.Error())
	}

	ciphertext := aesgcm.Seal(nil, ivBuf, dataBuf, nil)

	if c.Bool("export") {
		reader := bytes.NewReader(ciphertext)
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return errors.New("failed to copy to stdout: " + err.Error())
		}
		return nil
	}

	fmt.Fprintf(c.App.Writer, "%x\n", ciphertext)
	return nil
}

func decrypt(c *cli.Context) error {

	var dataBuf []byte
	var err error

	if c.Bool("readData") {
		dataBuf, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errors.New("failed to read data from stdin")
		}
	} else {
		data := c.String("data")
		if data == "" {
			return errors.New("please provide data with --data")
		}
		dataBuf, err = hex.DecodeString(data)
		if err != nil {
			return errors.New("failed to decode data string: " + err.Error())
		}
	}

	var key, iv string
	var keyBuf, ivBuf []byte

	keyAndInitVal := c.String("keyAndInitVal")
	if keyAndInitVal != "" {
		keyAndInitValBuf, err := hex.DecodeString(keyAndInitVal)
		if err != nil {
			return errors.New("failed to decode keyAndInitVal: " + err.Error())
		}
		// 16 bytes = 128 bits, 12 bytes = 96 bits
		if len(keyAndInitValBuf) != 16+12 {
			return fmt.Errorf("length of key + initVal must be 224 bits, not %d",
				len(keyAndInitValBuf)*8)
		}
		keyBuf = keyAndInitValBuf[:16]
		ivBuf = keyAndInitValBuf[16:]
	} else {
		key = c.String("key")
		if key == "" {
			return errors.New("please provide a key with --key")
		}

		iv = c.String("initVal")
		if iv == "" {
			return errors.New("please provice an initialization value with --initVal")
		}

		keyBuf, err = hex.DecodeString(key)
		if err != nil {
			return errors.New("failed to decode key as hexadecimal: " + err.Error())
		}
		if len(keyBuf) != 16 {
			return fmt.Errorf("length of key must be 128 bits, not %d",
				len(keyBuf)*8)
		}

		ivBuf, err = hex.DecodeString(iv)
		if err != nil {
			return errors.New("failed to decode initVal as hexadecimal: " + err.Error())
		}
		if len(ivBuf) != 12 {
			return fmt.Errorf("length of initialization value must be 96 bits, not %d",
				len(ivBuf)*8)
		}
	}

	block, err := aes.NewCipher(keyBuf)
	if err != nil {
		return errors.New("failed to create new cipher: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.New("failed to create new cipher: " + err.Error())
	}

	ciphertext, err := aesgcm.Open(nil, ivBuf, dataBuf, nil)
	if err != nil {
		return errors.New("failed to decode: " + err.Error())
	}

	if c.Bool("export") {
		reader := bytes.NewReader(ciphertext)
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return errors.New("failed to copy to stdout: " + err.Error())
		}
		return nil
	}

	fmt.Fprintf(c.App.Writer, "%s\n", string(ciphertext))
	return nil
}
