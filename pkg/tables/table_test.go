package tables_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Tables", func() {
	It("lists tables", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.Method).To(Equal(http.MethodGet))
			writeOCS(writer, http.StatusOK, []map[string]any{{"id": 42, "title": "Customers"}})
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		result, err := client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
	})

	It("creates, reads, updates and deletes tables", func(ctx SpecContext) {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPost:
				var body map[string]any
				Expect(json.NewDecoder(request.Body).Decode(&body)).To(Succeed())
				Expect(body["title"]).To(Equal("Customers"))
				Expect(body["template"]).To(Equal("custom"))
				writeOCS(writer, http.StatusOK, map[string]any{"id": 42, "title": "Customers"})
			case http.MethodGet:
				Expect(request.URL.Path).To(HaveSuffix("/tables/42"))
				writeOCS(writer, http.StatusOK, map[string]any{"id": 42, "title": "Customers"})
			case http.MethodPut:
				writeOCS(writer, http.StatusOK, map[string]any{"id": 42, "title": "Accounts"})
			case http.MethodDelete:
				writeOCS(writer, http.StatusOK, map[string]any{})
			default:
				Fail("unexpected HTTP method")
			}
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		created, err := client.CreateTable(ctx, tables.CreateTableOptions{Title: "Customers"})
		Expect(err).NotTo(HaveOccurred())
		fetched, err := client.GetTable(ctx, created.ID)
		Expect(err).NotTo(HaveOccurred())
		title := "Accounts"
		updated, err := client.UpdateTable(ctx, fetched.ID, tables.UpdateTableOptions{Title: &title})
		Expect(err).NotTo(HaveOccurred())
		err = client.DeleteTable(ctx, updated.ID)

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(created.ID).To(Equal(int64(42)))
		Expect(updated.Title).To(Equal("Accounts"))
	})

	It("creates lightweight immutable table scopes", func() {
		client, err := tables.NewClient("https://cloud.example.net")
		Expect(err).NotTo(HaveOccurred())
		first := client.Table(1)
		second := client.Table(2)
		Expect(first.ID()).To(Equal(int64(1)))
		Expect(second.ID()).To(Equal(int64(2)))
		Expect(first.Rows()).NotTo(Equal(second.Rows()))
		Expect(first.Columns()).NotTo(Equal(second.Columns()))
	})

	It("rejects invalid table arguments", func(ctx SpecContext) {
		client, err := tables.NewClient("https://cloud.example.net")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.CreateTable(ctx, tables.CreateTableOptions{})
		Expect(err).To(HaveOccurred())
		_, err = client.GetTable(ctx, 0)
		Expect(err).To(HaveOccurred())
		_, err = client.UpdateTable(ctx, 0, tables.UpdateTableOptions{})
		Expect(err).To(HaveOccurred())
		Expect(client.DeleteTable(ctx, 0)).To(HaveOccurred())
	})
})
