package tables_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Columns", func() {
	It("uses table-scoped column routes", func(ctx SpecContext) {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				Expect(request.URL.Path).To(HaveSuffix("/api/2/columns/table/42"))
				writeOCS(writer, http.StatusOK, []map[string]any{{"id": 7, "title": "Name", "type": "text"}})
			case http.MethodPost:
				Expect(request.URL.Path).To(HaveSuffix("/index.php/apps/tables/api/1/tables/42/columns"))
				var body map[string]any
				Expect(json.NewDecoder(request.Body).Decode(&body)).To(Succeed())
				Expect(body).To(HaveKeyWithValue("title", "Name"))
				Expect(body).To(HaveKey("numberDefault"))
				Expect(body["numberDefault"]).To(BeNil())
				Expect(body).To(HaveKey("textMaxLength"))
				writeJSON(writer, http.StatusOK, map[string]any{"id": 7, "title": "Name", "type": "text"})
			case http.MethodPut:
				Expect(request.URL.Path).To(HaveSuffix("/index.php/apps/tables/api/1/columns/7"))
				var body map[string]any
				Expect(json.NewDecoder(request.Body).Decode(&body)).To(Succeed())
				Expect(body).To(HaveKey("selectionOptions"))
				Expect(body["selectionOptions"]).To(BeNil())
				writeJSON(writer, http.StatusOK, map[string]any{"id": 7, "title": "Full name", "type": "text"})
			case http.MethodDelete:
				Expect(request.URL.Path).To(HaveSuffix("/index.php/apps/tables/api/1/columns/7"))
				writeJSON(writer, http.StatusOK, map[string]any{})
			}
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		columns := client.Table(42).Columns()

		By("Act")
		listed, err := columns.List(ctx, tables.ListColumnsOptions{})
		Expect(err).NotTo(HaveOccurred())
		created, err := columns.Create(ctx, tables.CreateColumnOptions{Title: "Name", Type: "text", Subtype: "line"})
		Expect(err).NotTo(HaveOccurred())
		newTitle := "Full name"
		updated, err := columns.Update(ctx, created.ID, tables.UpdateColumnOptions{Title: &newTitle})
		Expect(err).NotTo(HaveOccurred())
		err = columns.Delete(ctx, updated.ID)

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(HaveLen(1))
	})

	It("validates column arguments", func(ctx SpecContext) {
		client, err := tables.NewClient("https://cloud.example.net")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Table(0).Columns().List(ctx, tables.ListColumnsOptions{})
		Expect(err).To(HaveOccurred())
		_, err = client.Table(0).Columns().Create(ctx, tables.CreateColumnOptions{Title: "x", Type: "text"})
		Expect(err).To(HaveOccurred())
		_, err = client.Table(1).Columns().Create(ctx, tables.CreateColumnOptions{})
		Expect(err).To(HaveOccurred())
		_, err = client.Table(0).Columns().Update(ctx, 1, tables.UpdateColumnOptions{})
		Expect(err).To(HaveOccurred())
		_, err = client.Table(1).Columns().Update(ctx, 0, tables.UpdateColumnOptions{})
		Expect(err).To(HaveOccurred())
		Expect(client.Table(0).Columns().Delete(ctx, 1)).To(HaveOccurred())
		Expect(client.Table(1).Columns().Delete(ctx, 0)).To(HaveOccurred())
	})
})
