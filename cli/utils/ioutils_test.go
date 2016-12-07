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

package utils_test

import (
	"bytes"
	"fmt"
	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	. "github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("ioutils", func() {
	fmt.Println("")
	// TODO: Create test_files
	JSONPath := "test_file.json"
	YAMLPath := "test_file.yaml"

	var _ = BeforeSuite(func() {
		// Check JSON file
		json, JSONerr := ioutil.ReadFile(JSONPath)
		Expect(JSONerr).NotTo(HaveOccurred())
		Expect(json).NotTo(BeEmpty())

		// Check YAML file
		yaml, YAMLerr := ioutil.ReadFile(YAMLPath)
		Expect(YAMLerr).NotTo(HaveOccurred())
		Expect(yaml).NotTo(BeEmpty())

	})

	Describe("Parsing file", func() {

		Context("when the data type is not JSON or YAML", func() {
			It("should return an error", func() {
				_, format, err := ReadInputFile("test.txt")
				Expect(format).To(Equal("TXT"))
				Expect(err).To(HaveOccurred())
				Expect(err).Should(MatchError(common.ErrUnsoportedFormat))
			})
		})

		Context("when the data is JSON", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(JSONPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(JSON))
				rules := &api.Rule{}

				// Unmarshall JSON Reader
				err = UnmarshallReader(reader, format, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("json_id"))
				Expect(rules.Destination).To(Equal("json_destination"))

				// Marshall reader as JSON
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, format)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("\"id\": \"json_id\""))
			})

		})

		Context("when the data is YAML", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(YAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(YAML))
				rules := &api.Rule{}

				// Unmarshall YAML reader
				err = UnmarshallReader(reader, format, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				// Marshall reader as YAML
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, format)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("id: yaml_id"))
			})

		})

		Context("when converting from YAML to JSON", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(YAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(YAML))
				rules := &api.Rule{}

				// Convert YAML to JSON
				reader, err = YAMLToJSON(reader, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				// Unmarshall reader as JSON
				err = UnmarshallReader(reader, JSON, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				//Prettyfy as JSON
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, JSON)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("\"id\": \"yaml_id\""))
			})

		})
	})

})
