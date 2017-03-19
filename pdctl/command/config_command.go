// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	configPrefix    = "pd/api/v1/config"
	schedulePrefix  = "pd/api/v1/config/schedule"
	replicatePrefix = "pd/api/v1/config/replicate"
)

// NewConfigCommand return a config subcommand of rootCmd
func NewConfigCommand() *cobra.Command {
	conf := &cobra.Command{
		Use:   "config <subcommand>",
		Short: "tune pd configs",
	}
	conf.AddCommand(NewShowConfigCommand())
	conf.AddCommand(NewSetConfigCommand())
	return conf
}

// NewShowConfigCommand return a show subcommand of configCmd
func NewShowConfigCommand() *cobra.Command {
	sc := &cobra.Command{
		Use:   "show",
		Short: "show config of PD",
		Run:   showConfigCommandFunc,
	}
	sc.AddCommand(NewShowAllConfigCommand())
	return sc
}

// NewShowAllConfigCommand return a show all subcommand of show subcommand
func NewShowAllConfigCommand() *cobra.Command {
	sc := &cobra.Command{
		Use:   "all",
		Short: "show all config of PD",
		Run:   showAllConfigCommandFunc,
	}
	return sc
}

// NewSetConfigCommand return a set subcommand of configCmd
func NewSetConfigCommand() *cobra.Command {
	sc := &cobra.Command{
		Use:   "set <option> <value>",
		Short: "set the option with value",
		Run:   setConfigCommandFunc,
	}
	sc.AddCommand(NewSetReplicationCommand())
	return sc
}

// NewSetReplicationCommand return  a set replication subcommand of setCmd
func NewSetReplicationCommand() *cobra.Command {
	sc := &cobra.Command{
		Use:   "replicate <option> <value>",
		Short: "set the replication option with value",
		Run:   setReplicateConfigCommandFunc,
	}
	return sc
}

func showConfigCommandFunc(cmd *cobra.Command, args []string) {
	r, err := doRequest(cmd, schedulePrefix, http.MethodGet)
	if err != nil {
		fmt.Printf("Failed to get config: %s", err)
		return
	}
	fmt.Println(r)
}

func showAllConfigCommandFunc(cmd *cobra.Command, args []string) {
	r, err := doRequest(cmd, configPrefix, http.MethodGet)
	if err != nil {
		fmt.Printf("Failed to get config: %s", err)
		return
	}
	fmt.Println(r)
}

func setConfigCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println(cmd.UsageString())
		return
	}

	url := getAddressFromCmd(cmd, schedulePrefix)
	var value interface{}
	data := make(map[string]interface{})

	r, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to set config:[%s]\n", err)
		return
	}
	if r.StatusCode != http.StatusOK {
		printResponseError(r)
		r.Body.Close()
		return
	}

	json.NewDecoder(r.Body).Decode(&data)
	r.Body.Close()
	value, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		value = args[1]
	}
	data[args[0]] = value

	req, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to set config:[%s]\n", err)
		return
	}

	url = getAddressFromCmd(cmd, configPrefix)
	r, err = http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		fmt.Printf("Failed to set config:[%s]\n", err)
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		fmt.Println("Success!")
	} else {
		printResponseError(r)
	}
}

func setReplicateConfigCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println(cmd.UsageString())
		return
	}

	var value interface{}
	data := make(map[string]interface{})
	r, err := doRequest(cmd, replicatePrefix, http.MethodGet)
	if err != nil {
		fmt.Printf("Failed to set replication config:[%s]\n", err)
		return
	}
	err = json.Unmarshal([]byte(r), &data)
	if err != nil {
		fmt.Printf("Failed to set replication config:[%s]\n", err)
		return
	}
	value, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		value = args[1]
		if strings.Contains(args[1], ",") {
			value = strings.Split(args[1], ",")
		}
	}
	data[args[0]] = value
	reqData, err := json.Marshal(&data)
	if err != nil {
		fmt.Printf("Failed to set replication config:[%s]\n", err)
		return
	}
	url := getAddressFromCmd(cmd, replicatePrefix)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		fmt.Printf("Failed to set replication config:[%s]\n", err)
	}
	defer func() {
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Success!")
	} else {
		printResponseError(resp)
	}
}
