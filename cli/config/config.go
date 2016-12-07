package config

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/urfave/cli"
)

// Before runs after the context is ready and before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func Before(ctx *cli.Context) error {
	return nil
}

// OnUsageError .
func OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	if err != nil {
		logrus.WithError(err).Debug("Error")

		if strings.Contains(err.Error(), common.ErrUnknowFlag.Error()) {
			cli.ShowAppHelp(ctx)
			return nil
		}

		if strings.Contains(err.Error(), common.ErrInvalidFlagArg.Error()) {
			flag := err.Error()[strings.LastIndex(err.Error(), "-")+1:]

			if flag == common.RegistryURL.Flag() {
				_, err = api.ValidateRegistryURL(ctx)
				if err != nil {
					fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
					return nil
				}
			}

			if flag == common.ControllerURL.Flag() {
				_, err = api.ValidateControllerURL(ctx)
				if err != nil {
					fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
					return nil
				}
			}
		}

		fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
		return err
	}

	cli.ShowAppHelp(ctx)
	return nil
}

// DefaultAction .
func DefaultAction(ctx *cli.Context) error {
	// Validate flags if not command has been specified
	if ctx.NumFlags() > 0 && ctx.NArg() == 0 {
		_, err := api.ValidateRegistryURL(ctx)
		if err != nil {
			fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
			return nil
		}

		_, err = api.ValidateControllerURL(ctx)
		if err != nil {
			fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
			return nil
		}
	}

	cli.ShowAppHelp(ctx)
	return nil
}
