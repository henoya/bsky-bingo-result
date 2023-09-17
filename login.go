package main

import (
	"context"
	"encoding/json"
	"fmt"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
)

func execLogin(cCtx *cli.Context, host string, handle string, passwd string) (data []byte, err error) {
	fp, _ := cCtx.App.Metadata["path"].(string)
	var cfg config
	cfg.Host = host
	cfg.Handle = handle
	cfg.Password = passwd
	b, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cannot make config file: %w", err)
	}
	err = ioutil.WriteFile(fp, b, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot write config file: %w", err)
	}

	return b, nil
}

func doLogin(cCtx *cli.Context) error {
	host := cCtx.String("host")
	handle := cCtx.Args().Get(0)
	password := cCtx.Args().Get(1)
	if handle == "" || password == "" {
		cli.ShowSubcommandHelpAndExit(cCtx, 1)
	}
	_, err := execLogin(cCtx, host, handle, password)
	if err != nil {
		return fmt.Errorf("cannot make config file: %w", err)
	}
	return nil
}

func execSession(cCtx *cli.Context) (session *comatproto.ServerGetSession_Output, err error) {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}

	session, err = comatproto.ServerGetSession(context.TODO(), xrpcc)
	if err != nil {
		return nil, err
	}

	if cCtx.Bool("json") {
		json.NewEncoder(os.Stdout).Encode(session)
		return session, nil
	}
	return session, nil
}

func doShowSession(cCtx *cli.Context) error {
	session, err := execSession(cCtx)
	if err != nil {
		return err
	}
	fmt.Printf("Did: %s\n", session.Did)
	fmt.Printf("Email: %s\n", stringp(session.Email))
	fmt.Printf("Handle: %s\n", session.Handle)
	return nil
}
