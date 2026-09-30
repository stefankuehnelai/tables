package tables_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Rows", func() {
	It("uses one row API for authenticated table CRUD", func(ctx SpecContext) {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				Expect(request.URL.Path).To(HaveSuffix("/index.php/apps/tables/api/1/tables/42/rows"))
				Expect(request.URL.Query().Get("limit")).To(Equal("100"))
				Expect(request.URL.Query().Get("offset")).To(Equal("5"))
				writeJSON(writer, http.StatusOK, []map[string]any{{"id": 5, "tableId": 42, "data": []map[string]any{{"columnId": 12, "value": "Alice"}}}})
			case http.MethodPost:
				Expect(request.URL.Path).To(HaveSuffix("/api/2/tables/42/rows"))
				var body struct {
					Data map[string]any `json:"data"`
				}
				Expect(json.NewDecoder(request.Body).Decode(&body)).To(Succeed())
				Expect(body.Data["12"]).To(Equal("Alice"))
				writeOCS(writer, http.StatusOK, map[string]any{"id": 5, "tableId": 42})
			case http.MethodPut:
				Expect(request.URL.Path).To(HaveSuffix("/api/2/tables/42/rows/5"))
				writeOCS(writer, http.StatusOK, map[string]any{"id": 5, "tableId": 42})
			case http.MethodDelete:
				Expect(request.URL.Path).To(HaveSuffix("/api/2/tables/42/rows/5"))
				writeOCS(writer, http.StatusOK, map[string]any{})
			}
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := client.Table(42).Rows()

		By("Act")
		listed, err := rows.List(ctx, tables.ListRowsOptions{Limit: 100, Offset: 5})
		Expect(err).NotTo(HaveOccurred())
		created, err := rows.Create(ctx, tables.CreateRowOptions{Values: tables.RowValues{12: "Alice"}})
		Expect(err).NotTo(HaveOccurred())
		updated, err := rows.Update(ctx, created.ID, tables.UpdateRowOptions{Values: tables.RowValues{12: "Alicia"}})
		Expect(err).NotTo(HaveOccurred())
		err = rows.Delete(ctx, updated.ID)

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(listed[0].Value(12)).To(Equal("Alice"))
		Expect(listed[0].Value(999)).To(BeNil())
	})

	It("validates row options and scopes", func(ctx SpecContext) {
		client, err := tables.NewClient("https://cloud.example.net")
		Expect(err).NotTo(HaveOccurred())
		rows := client.Table(42).Rows()
		_, err = rows.List(ctx, tables.ListRowsOptions{Limit: -1})
		Expect(err).To(HaveOccurred())
		_, err = rows.List(ctx, tables.ListRowsOptions{Offset: -1})
		Expect(err).To(HaveOccurred())
		_, err = rows.Create(ctx, tables.CreateRowOptions{})
		Expect(err).To(HaveOccurred())
		_, err = rows.Update(ctx, 0, tables.UpdateRowOptions{Values: tables.RowValues{1: "x"}})
		Expect(err).To(HaveOccurred())
		_, err = rows.Update(ctx, 1, tables.UpdateRowOptions{})
		Expect(err).To(HaveOccurred())
		Expect(rows.Delete(ctx, 0)).To(HaveOccurred())
		_, err = client.Table(0).Rows().List(ctx, tables.ListRowsOptions{})
		Expect(err).To(HaveOccurred())
		var empty tables.Rows
		_, err = empty.List(ctx, tables.ListRowsOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeInvalidConfiguration))
	})
})
