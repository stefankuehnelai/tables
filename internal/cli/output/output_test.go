package output_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/internal/cli/output"
)

var _ = Describe("Formatter", func() {
	It("validates structured output combinations", func() {
		_, jqErr := output.New(output.Options{JQ: "."})
		_, templateErr := output.New(output.Options{Template: "{{.}}"})
		_, bothErr := output.New(output.Options{JSONFields: []string{"id"}, JQ: ".", Template: "{{.}}"})
		Expect(jqErr).To(MatchError("--jq requires --json"))
		Expect(templateErr).To(MatchError("--template requires --json"))
		Expect(bothErr).To(MatchError("--jq and --template are mutually exclusive"))
	})

	It("writes JSON and applies pure-Go jq and templates", func() {
		value := []map[string]any{{"id": 7, "title": "Customers"}, {"id": 8, "title": "Other"}}
		formatter, err := output.New(output.Options{JSONFields: []string{"id", "title"}, JQ: `.[] | select(.title == "Customers") | .id`})
		Expect(err).NotTo(HaveOccurred())
		var jq bytes.Buffer
		Expect(formatter.Write(&jq, value)).To(Succeed())
		Expect(jq.String()).To(Equal("7\n"))

		formatter, err = output.New(output.Options{JSONFields: []string{"id", "title"}, Template: `{{range .}}{{.id}} {{.title}}{{"\n"}}{{end}}`})
		Expect(err).NotTo(HaveOccurred())
		var tmpl bytes.Buffer
		Expect(formatter.Write(&tmpl, value[:1])).To(Succeed())
		Expect(tmpl.String()).To(Equal("7 Customers\n"))
	})

	It("rejects unknown fields and renders human tables", func() {
		formatter, err := output.New(output.Options{JSONFields: []string{"missing"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(formatter.Write(&bytes.Buffer{}, map[string]any{"id": 1})).To(MatchError(`unknown JSON field "missing"`))

		formatter, err = output.New(output.Options{})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer
		Expect(formatter.Write(&buffer, []map[string]any{{"title": "Customers", "id": 4, "nested": map[string]any{"ignored": true}}})).To(Succeed())
		Expect(buffer.String()).To(ContainSubstring("Customers"))
		Expect(buffer.String()).NotTo(ContainSubstring("nested"))
	})
})
