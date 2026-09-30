package output_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/internal/cli/output"
)

var _ = Describe("Formatter", func() {
	It("validates structured output combinations", func() {
		By("Arrange and Act")
		_, jqErr := output.New(output.Options{JQ: "."})
		_, templateErr := output.New(output.Options{Template: "{{.}}"})
		_, bothErr := output.New(output.Options{JSONFields: []string{"id"}, JQ: ".", Template: "{{.}}"})

		By("Assert")
		Expect(jqErr).To(MatchError("--jq requires --json"))
		Expect(templateErr).To(MatchError("--template requires --json"))
		Expect(bothErr).To(MatchError("--jq and --template are mutually exclusive"))
	})

	It("writes JSON with selected fields", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"id", "title"}})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer

		By("Act")
		err = formatter.Write(&buffer, []map[string]any{{"ID": 1, "Title": "Customers", "Ignored": true}})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(MatchJSON(`[{"id":1,"title":"Customers"}]`))
	})

	It("rejects unknown JSON fields", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"missing"}})
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		err = formatter.Write(&bytes.Buffer{}, map[string]any{"id": 1})

		By("Assert")
		Expect(err).To(MatchError(`unknown JSON field "missing"`))
	})

	It("rejects empty and non-object selections", func() {
		By("Arrange")
		empty, err := output.New(output.Options{JSONFields: []string{" "}})
		Expect(err).NotTo(HaveOccurred())
		scalar, err := output.New(output.Options{JSONFields: []string{"id"}})
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		emptyErr := empty.Write(&bytes.Buffer{}, map[string]any{"id": 1})
		scalarErr := scalar.Write(&bytes.Buffer{}, "value")

		By("Assert")
		Expect(emptyErr).To(MatchError("JSON field name must not be empty"))
		Expect(scalarErr).To(MatchError("JSON field selection requires an object or array of objects"))
	})

	It("evaluates jq without an external process", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"id", "title"}, JQ: `.[] | select(.title == "Customers") | .id`})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer

		By("Act")
		err = formatter.Write(&buffer, []map[string]any{{"id": 7, "title": "Customers"}, {"id": 8, "title": "Other"}})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(Equal("7\n"))
	})

	It("reports invalid jq", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"id"}, JQ: ".["})
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		err = formatter.Write(&bytes.Buffer{}, map[string]any{"id": 7})

		By("Assert")
		Expect(err).To(MatchError(ContainSubstring("evaluate jq expression")))
	})

	It("renders templates", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"id", "title"}, Template: `{{range .}}{{.id}} {{.title}}{{"\n"}}{{end}}`})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer

		By("Act")
		err = formatter.Write(&buffer, []map[string]any{{"id": 1, "title": "Customers"}})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(Equal("1 Customers\n"))
	})

	It("reports invalid templates", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{JSONFields: []string{"id"}, Template: "{{"})
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		err = formatter.Write(&bytes.Buffer{}, map[string]any{"id": 1})

		By("Assert")
		Expect(err).To(MatchError(ContainSubstring("parse template")))
	})

	It("renders a stable human table", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer

		By("Act")
		err = formatter.Write(&buffer, []map[string]any{{"title": "Customers", "id": 4, "nested": map[string]any{"ignored": true}}})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(ContainSubstring("ID"))
		Expect(buffer.String()).To(ContainSubstring("TITLE"))
		Expect(buffer.String()).To(ContainSubstring("Customers"))
		Expect(buffer.String()).NotTo(ContainSubstring("nested"))
	})

	It("renders empty result without output", func() {
		By("Arrange")
		formatter, err := output.New(output.Options{})
		Expect(err).NotTo(HaveOccurred())
		var buffer bytes.Buffer

		By("Act")
		err = formatter.Write(&buffer, []map[string]any{})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(BeEmpty())
	})

	It("rejects scalar table output", func() {
		formatter, err := output.New(output.Options{})
		Expect(err).NotTo(HaveOccurred())
		err = formatter.Write(&bytes.Buffer{}, "hello")
		Expect(err).To(MatchError(ContainSubstring("table output requires an object or array")))
	})
})
