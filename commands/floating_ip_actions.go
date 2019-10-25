/*
Copyright 2018 The Doctl Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"fmt"
	"strconv"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/commands/displayers"
	"github.com/digitalocean/doctl/do"
	"github.com/spf13/cobra"
)

// FloatingIPAction creates the floating IP action command.
func FloatingIPAction() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:     "floating-ip-action",
			Short:   "Provides commands to perform actions on Floating IP Addresses",
			Long:    "Floating IP actions are commands that can be given to a DigitalOcean Floating IP.",
			Aliases: []string{"fipa"},
		},
	}
	flipactionDetail := `

	- The unique numeric ID used to identify and reference a Floating IP action. 
	- The status of the Floating IP action. This will be either "in-progress", "completed", or "errored".
	- A time value given in ISO8601 combined date and time format that represents when the action was initiated.
	- A time value given in ISO8601 combined date and time format that represents when the action was completed.
	- The resource ID, which is a unique identifier for the resource that the action is associated with.
	- The type of resource that the action is associated with.
	- The region where the action occurred.
	- The slug for the region where the action occurred.
`
	CmdBuilderWithDocs(cmd, RunFloatingIPActionsGet,
		"get <floating-ip> <action-id>", "Retrive the status of a Floating IP action",`Use this command to retrieve the status of a Floating IP action. Outputs the following information:`+flipactionDetail, Writer,
		displayerType(&displayers.Action{}))

	CmdBuilderWithDocs(cmd, RunFloatingIPActionsAssign,
		"assign <floating-ip> <droplet-id>", "Assign a Floating IP Address to a Droplet",`Use this command to assign a Floating IP Address to a Droplet. Set the "droplet_id" attribute to the Droplet's ID.`, Writer,
		displayerType(&displayers.Action{}))

	CmdBuilderWithDocs(cmd, RunFloatingIPActionsUnassign,
		"unassign <floating-ip>", "Unassign a Floating IP Address from a Droplet",`Use this command to unassign a Floating IP Address. The Floating IP Address will be reserved in the region but not assigned to a Droplet.`, Writer,
		displayerType(&displayers.Action{}))

	return cmd
}

// RunFloatingIPActionsGet retrieves an action for a floating IP.
func RunFloatingIPActionsGet(c *CmdConfig) error {
	if len(c.Args) != 2 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	ip := c.Args[0]

	fia := c.FloatingIPActions()

	actionID, err := strconv.Atoi(c.Args[1])
	if err != nil {
		return err
	}

	a, err := fia.Get(ip, actionID)
	if err != nil {
		return err
	}

	item := &displayers.Action{Actions: do.Actions{*a}}
	return c.Display(item)
}

// RunFloatingIPActionsAssign assigns a floating IP to a droplet.
func RunFloatingIPActionsAssign(c *CmdConfig) error {
	if len(c.Args) != 2 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	ip := c.Args[0]

	fia := c.FloatingIPActions()

	dropletID, err := strconv.Atoi(c.Args[1])
	if err != nil {
		return err
	}

	a, err := fia.Assign(ip, dropletID)
	if err != nil {
		checkErr(fmt.Errorf("could not assign IP to droplet: %v", err))
	}

	item := &displayers.Action{Actions: do.Actions{*a}}
	return c.Display(item)
}

// RunFloatingIPActionsUnassign unassigns a floating IP to a droplet.
func RunFloatingIPActionsUnassign(c *CmdConfig) error {
	if len(c.Args) != 1 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	ip := c.Args[0]

	fia := c.FloatingIPActions()

	a, err := fia.Unassign(ip)
	if err != nil {
		checkErr(fmt.Errorf("could not unassign IP to droplet: %v", err))
	}

	item := &displayers.Action{Actions: do.Actions{*a}}
	return c.Display(item)
}
