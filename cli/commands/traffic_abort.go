// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package commands

import (
	"bytes"
	"strings"

	"fmt"

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// TrafficAbortCommand is used for the route-list command.
type TrafficAbortCommand struct {
	ctx        *cli.Context
	controller api.ControllerClient
	term       terminal.UI
}

// NewTrafficAbortCommand constructs a new TrafficAbort.
func NewTrafficAbortCommand(term terminal.UI) (cmd *TrafficAbortCommand) {
	return &TrafficAbortCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *TrafficAbortCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("traffic_abort_name"),
		Description: T("traffic_abort_description"),
		Usage:       T("traffic_abort_usage"),
		// TODO: Complete UsageText
		UsageText: T("traffic_abort_name"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: T("traffic_abort_service_usage"),
				Value: "",
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *TrafficAbortCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *TrafficAbortCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *TrafficAbortCommand) Action(ctx *cli.Context) error {
	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.IsSet("service") {
		return cmd.AbortTraffic(ctx.String("service"))
	}

	if ctx.NArg() > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	return cmd.DefaultAction(ctx)
}

// AbortTraffic .
func (cmd *TrafficAbortCommand) AbortTraffic(serviceName string) error {
	routes, err := cmd.controller.ServiceRoutes(serviceName)
	if err != nil {
		return err
	}

	if len(routes.Rules) == 0 {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s: %q\n\n", common.ErrNotRulesFoundForService.Error(), serviceName)
		return nil
	}

	if len(routes.Rules) > 1 || len(routes.Rules[0].Route.Backends) != 2 || routes.Rules[0].Route.Backends[0].Weight == routes.Rules[0].Route.Backends[1].Weight {
		fmt.Fprintf(cmd.ctx.App.Writer, "Invalid state for step operation\n\n")
		return nil
	}

	rule := routes.Rules[0]

	// Sort backends by weight, make sure default is last in the slice
	if rule.Route.Backends[0].Weight == 0 {
		rule.Route.Backends[0], rule.Route.Backends[1] = rule.Route.Backends[1], rule.Route.Backends[0]
	}

	defaultVersion := strings.Join(rule.Route.Backends[1].Tags, ", ")
	rule.Route.Backends = []api.Backend{rule.Route.Backends[1]}

	ruleList := api.RuleList{
		Rules: []api.Rule{
			rule,
		},
	}

	buf := bytes.Buffer{}
	err = utils.MarshallReader(&buf, &ruleList, JSON)

	if err != nil {
		return err
	}
	payload := bytes.NewReader(buf.Bytes())

	_, err = cmd.controller.UpdateRules(payload)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.ctx.App.Writer, "Transfer aborted for %q: all traffic reverted to %q\n\n", serviceName, defaultVersion)
	return nil
}

// DefaultAction runs the default action.
func (cmd *TrafficAbortCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}
