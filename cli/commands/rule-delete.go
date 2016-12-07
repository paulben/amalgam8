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
	"fmt"

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// RuleDeleteCommand is used for the rule-delete command.
type RuleDeleteCommand struct {
	ctx        *cli.Context
	controller api.ControllerClient
	term       terminal.UI
}

// NewRuleDeleteCommand constructs a new Rule Delete command.
func NewRuleDeleteCommand(term terminal.UI) (cmd *RuleDeleteCommand) {
	return &RuleDeleteCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *RuleDeleteCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("rule_delete_name"),
		Description: T("rule_delete_description"),
		Usage:       T("rule_delete_usage"),
		// TODO: Complete UsageText
		UsageText: T("rule_delete_name"),
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "id, i",
				Usage: T("rule_delete_id_usage"),
			},
			cli.StringSliceFlag{
				Name:  "tag, t",
				Usage: T("rule_delete_tag_usage"),
			},
			cli.StringSliceFlag{
				Name:  "destination, d",
				Usage: T("rule_delete_destination_usage"),
			},
			cli.BoolFlag{
				Name:  "all, a",
				Usage: T("rule_delete_all_usage"),
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: T("rule_delete_all_force_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RuleDeleteCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RuleDeleteCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RuleDeleteCommand) Action(ctx *cli.Context) error {
	T := utils.Language(common.DefaultLanguage)
	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.IsSet("all") {
		switch ctx.Bool("force") {
		case true:
			return cmd.DeleteRules("")
		case false:
			confirmation, err := utils.Confirmation(ctx.App.Writer, T("rule_delete_all_confirmation"))
			if err != nil {
				return err
			}
			if confirmation {
				return cmd.DeleteRules("")
			}
			return nil
		}
	}

	query := cmd.controller.NewQuery()
	if ctx.IsSet("id") || ctx.IsSet("i") || ctx.IsSet("destination") || ctx.IsSet("d") || ctx.IsSet("tag") || ctx.IsSet("t") {
		for _, id := range ctx.StringSlice("id") {
			query.Add("id", id)
		}
		for _, dest := range ctx.StringSlice("destination") {
			query.Add("destination", dest)
		}
		for _, dest := range ctx.StringSlice("tag") {
			query.Add("tags", dest)
		}
		return cmd.DeleteRules(query.Encode())
	}

	return cmd.DefaultAction(ctx)
}

// DeleteRules deletes the rules based on the given query.
func (cmd *RuleDeleteCommand) DeleteRules(query string) error {
	T := utils.Language(common.DefaultLanguage)
	_, err := cmd.controller.DeleteRules(query)
	if err != nil {
		return err
	}
	// TODO: Add errors in client.
	fmt.Fprintln(cmd.ctx.App.Writer, T("request_completed"))
	return nil
}

// DefaultAction runs the default action.
func (cmd *RuleDeleteCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}
