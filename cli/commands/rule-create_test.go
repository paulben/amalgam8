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

package commands_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	cmds "github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/flags"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/urfave/cli"
)

var _ = Describe("Rule-Create", func() {
	fmt.Println()
	utils.LoadLocales()
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.RuleCreateCommand
	var app *cli.App
	var server *ghttp.Server
	response := make(map[string][]byte)

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = flags.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewRuleCreateCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Setup()

		response["json_id"] = []byte(`
			{
  			"ids": [
    		"json_id"
			  ]
			}`)
		response["yaml_id"] = []byte(`
				{
	  			"ids": [
	    		"yaml_id"
				  ]
				}`)
	})

	Describe("Create Rules", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "rule-create"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=", "rule-create"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).To(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Incorrect Usage"))
			})

		})

		Describe("On usage error: [rule-create bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, ""),
					),
				)
			})

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-create", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Create rules using a file: [rule-create -f]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["json_id"]),
					),
				)
			})

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should read rules from the JSON file", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--debug=true", "rule-create", "-f=fake_rule.json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"json_id"`))
				fmt.Println(app.Writer)
			})

		})

		Describe("Create rules using a file: [rule-create -f]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["yaml_id"]),
					),
				)
			})

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should read rules from the YAML file", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--debug=true", "rule-create", "-f=fake_rule.yaml"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"yaml_id"`))
				fmt.Println(app.Writer)
			})

		})

	})
})
