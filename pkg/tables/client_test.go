package tables_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

func writeOCS(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	Expect(json.NewEncoder(writer).Encode(map[string]any{
		"ocs": map[string]any{
			"meta": map[string]any{"status": "ok", "statuscode": status, "message": "OK"},
			"data": data,
		},
	})).To(Succeed())
}

func writeJSON(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	Expect(json.NewEncoder(writer).Encode(data)).To(Succeed())
}

var _ = Describe("Client", func() {
	It("validates and normalizes server URLs", func() {
		By("Act")
		client, err := tables.NewClient("https://cloud.example.net/nextcloud/?ignored=1#fragment")

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(client.BaseURL().String()).To(Equal("https://cloud.example.net/nextcloud"))
		Expect((*tables.Client)(nil).BaseURL()).To(BeNil())
	})

	DescribeTable("rejects invalid URLs", func(rawURL string) {
		_, err := tables.NewClient(rawURL)
		Expect(err).To(HaveOccurred())
	},
		Entry("missing scheme", "cloud.example.net"),
		Entry("unsupported scheme", "ftp://cloud.example.net"),
		Entry("missing host", "https:///nextcloud"),
		Entry("invalid escape", "https://example.net/%zz"),
	)

	It("sends OCS headers, query parameters, JSON, and basic authentication", func(ctx SpecContext) {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(Equal("/nextcloud/ocs/v2.php/apps/tables/api/2/tables"))
			Expect(request.URL.Query().Get("format")).To(Equal("json"))
			Expect(request.Header.Get("Accept")).To(Equal("application/json"))
			Expect(request.Header.Get("OCS-APIRequest")).To(Equal("true"))
			username, password, ok := request.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("alice"))
			Expect(password).To(Equal("secret"))
			writeOCS(writer, http.StatusOK, []any{})
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL+"/nextcloud", tables.WithCredentials("alice", "secret"))
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("maps HTTP errors to stable codes", func(status int, code int) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(status)
			_, _ = writer.Write([]byte("failure"))
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.GetTable(context.Background(), 42)
		var tablesErr tables.TablesError
		Expect(errors.As(err, &tablesErr)).To(BeTrue())
		Expect(tablesErr.Code()).To(Equal(code))
	},
		Entry("bad request", http.StatusBadRequest, tables.CodeInvalidArgument),
		Entry("unauthorized", http.StatusUnauthorized, tables.CodeUnauthorized),
		Entry("forbidden", http.StatusForbidden, tables.CodeForbidden),
		Entry("not found", http.StatusNotFound, tables.CodeNotFound),
		Entry("conflict", http.StatusConflict, tables.CodeConflict),
		Entry("server", http.StatusInternalServerError, tables.CodeServer),
		Entry("other", http.StatusTeapot, tables.CodeUnknown),
	)

	It("uses OCS error messages", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"ocs":{"meta":{"status":"failure","statuscode":404,"message":"missing"},"data":[]}}`))
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.GetTable(ctx, 42)
		Expect(err).To(MatchError("missing"))
	})

	It("reports malformed and incomplete OCS responses", func(ctx SpecContext) {
		responses := []string{"not-json", `{"ocs":{"meta":{"status":"ok","statuscode":200},"data":null}}`, `{"ocs":{"meta":{"status":"ok","statuscode":200},"data":"wrong"}}`}
		for _, payload := range responses {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				_, _ = writer.Write([]byte(payload))
			}))
			client, err := tables.NewClient(server.URL)
			Expect(err).NotTo(HaveOccurred())
			_, err = client.ListTables(ctx, tables.ListTablesOptions{})
			Expect(err).To(HaveOccurred())
			server.Close()
		}
	})

	It("maps OCS failures independently of HTTP status", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_, _ = writer.Write([]byte(`{"ocs":{"meta":{"status":"failure","statuscode":403,"message":"denied"},"data":[]}}`))
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeForbidden))
	})

	It("reports transport failures", func(ctx SpecContext) {
		client, err := tables.NewClient("http://127.0.0.1:1", tables.WithHTTPClient(&http.Client{}))
		Expect(err).NotTo(HaveOccurred())
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeTransport))
	})

	It("reports response read failures", func(ctx SpecContext) {
		client, err := tables.NewClient("https://example.invalid", tables.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: failingReader{}}, nil
		})}))
		Expect(err).NotTo(HaveOccurred())
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeInvalidResponse))
	})

	It("decodes regular JSON endpoints and their error messages", func(ctx SpecContext) {
		By("Arrange")
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests++
			Expect(request.URL.Path).To(Equal("/index.php/apps/tables/api/1/tables/42/rows"))
			Expect(request.Header.Get("OCS-APIRequest")).To(BeEmpty())
			if requests == 1 {
				writeJSON(writer, http.StatusOK, []map[string]any{{"id": 5, "tableId": 42}})
				return
			}
			writeJSON(writer, http.StatusForbidden, map[string]any{"message": "denied"})
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		rows, err := client.Table(42).Rows().List(ctx, tables.ListRowsOptions{})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(1))
		_, err = client.Table(42).Rows().List(ctx, tables.ListRowsOptions{})
		Expect(err).To(MatchError("denied"))
	})

	It("accepts successful requests without an output body", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.Method).To(Equal(http.MethodDelete))
			writeOCS(writer, http.StatusOK, map[string]any{})
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(client.DeleteTable(ctx, 1)).To(Succeed())
	})
})

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (failingReader) Close() error             { return nil }

var _ io.ReadCloser = failingReader{}
