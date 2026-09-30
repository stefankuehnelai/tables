package tables

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rows", func() {
	It("uses one CRUD API for authenticated table rows", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			data := `{"id":9,"tableId":42,"data":[{"columnId":7,"value":"Alice"}]}`
			if request.Method == http.MethodGet {
				Expect(request.URL.Path).To(Equal("/index.php/apps/tables/api/1/tables/42/rows"))
				Expect(request.URL.Query().Get("limit")).To(Equal("10"))
				Expect(request.URL.Query().Get("offset")).To(Equal("2"))
				_, _ = io.WriteString(writer, `[`+data+`]`)
				return
			}
			Expect(request.URL.Path).To(HavePrefix("/ocs/v2.php/apps/tables/api/2/tables/42/rows"))
			if request.Method == http.MethodPost || request.Method == http.MethodPut {
				var body struct {
					Data map[string]any `json:"data"`
				}
				Expect(json.NewDecoder(request.Body).Decode(&body)).To(Succeed())
				Expect(body.Data).To(HaveKey("7"))
			}
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":`+data+`}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := client.Table(42).Rows()

		list, err := rows.List(ctx, ListRowsOptions{Limit: 10, Offset: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		value, ok := list[0].Value(7)
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal("Alice"))
		_, ok = list[0].Value(8)
		Expect(ok).To(BeFalse())
		created, err := rows.Create(ctx, CreateRowOptions{Values: RowValues{7: "Alice"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.ID).To(Equal(int64(9)))
		updated, err := rows.Update(ctx, 9, UpdateRowOptions{Values: RowValues{7: "Bob"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.ID).To(Equal(int64(9)))
		deleted, err := rows.Delete(ctx, 9)
		Expect(err).NotTo(HaveOccurred())
		Expect(deleted.ID).To(Equal(int64(9)))
	})

	It("validates row options", func(ctx SpecContext) {
		client, err := NewClient("https://example.test")
		Expect(err).NotTo(HaveOccurred())
		rows := client.Table(42).Rows()
		_, err = rows.List(ctx, ListRowsOptions{Limit: -1})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = rows.Create(ctx, CreateRowOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = rows.Create(ctx, CreateRowOptions{Values: RowValues{0: "x"}})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = rows.Update(ctx, 0, UpdateRowOptions{Values: RowValues{1: "x"}})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.Table(0).Rows().Delete(ctx, 1)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		payload := rowPayload(RowValues{2: "b", 1: "a"})
		Expect(payload["data"]).To(Equal(RowValues{2: "b", 1: "a"}))
	})

	It("keeps public row listing on the OCS route", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(Equal("/ocs/v2.php/apps/tables/api/2/public/abcdefghijklmnop/rows"))
			Expect(strings.Contains(request.URL.RawQuery, "format=json")).To(BeTrue())
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":[]}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(Share{Token: "abcdefghijklmnop"}).Rows().List(ctx, ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
})
