package tables

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tables", func() {
	It("performs table CRUD and creates scopes", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			data := `{"id":42,"title":"Customers"}`
			if request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/tables") {
				data = `[{"id":42,"title":"Customers"}]`
			}
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":`+data+`}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())

		list, err := client.ListTables(ctx, ListTablesOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		got, err := client.GetTable(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Title).To(Equal("Customers"))
		created, err := client.CreateTable(ctx, CreateTableOptions{Title: "Customers"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.ID).To(Equal(int64(42)))
		title := "People"
		updated, err := client.UpdateTable(ctx, 42, UpdateTableOptions{Title: &title})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.ID).To(Equal(int64(42)))
		deleted, err := client.DeleteTable(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(deleted.ID).To(Equal(int64(42)))
		Expect(client.Table(42).Rows().tableID).To(Equal(int64(42)))
		Expect(client.Table(42).Columns().tableID).To(Equal(int64(42)))
	})

	It("validates table arguments", func(ctx SpecContext) {
		client, err := NewClient("https://example.test")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.GetTable(ctx, 0)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.CreateTable(ctx, CreateTableOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.UpdateTable(ctx, -1, UpdateTableOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.DeleteTable(ctx, 0)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
	})
})
