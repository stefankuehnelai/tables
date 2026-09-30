package tables

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Columns", func() {
	It("performs column CRUD through the scoped API", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(HavePrefix("/index.php/apps/tables/api/1/"))
			writer.Header().Set("Content-Type", "application/json")
			if request.Method == http.MethodGet {
				_, _ = io.WriteString(writer, `[{"id":7,"tableId":42,"title":"Name","type":"text"}]`)
				return
			}
			_, _ = io.WriteString(writer, `{"id":7,"tableId":42,"title":"Name","type":"text"}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		columns := client.Table(42).Columns()

		list, err := columns.List(ctx, ListColumnsOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		created, err := columns.Create(ctx, CreateColumnOptions{Title: "Name", Type: "text"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.ID).To(Equal(int64(7)))
		title := "Display name"
		updated, err := columns.Update(ctx, 7, UpdateColumnOptions{Title: &title})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.TableID).To(Equal(int64(42)))
		deleted, err := columns.Delete(ctx, 7)
		Expect(err).NotTo(HaveOccurred())
		Expect(deleted.ID).To(Equal(int64(7)))
	})

	It("validates scoped identifiers and creation fields", func(ctx SpecContext) {
		client, err := NewClient("https://example.test")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Table(0).Columns().List(ctx, ListColumnsOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.Table(42).Columns().Create(ctx, CreateColumnOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.Table(42).Columns().Update(ctx, 0, UpdateColumnOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.Table(0).Columns().Delete(ctx, 1)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
	})

	It("builds explicit nullable payloads", func() {
		payload := createColumnPayload(42, CreateColumnOptions{Title: "Name", Type: "text"})
		Expect(payload["tableId"]).To(Equal(int64(42)))
		Expect(payload["technicalName"]).To(BeNil())
		Expect(nullableString("")).To(BeNil())
		Expect(nullableString("x")).To(Equal("x"))
		title := "x"
		Expect(updateColumnPayload(UpdateColumnOptions{Title: &title})["title"]).To(Equal(&title))
		Expect(strings.TrimSpace(" x ")).To(Equal("x"))
	})
})
