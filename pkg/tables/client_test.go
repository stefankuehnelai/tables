package tables

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

var _ = Describe("Client", func() {
	It("normalizes subpath URLs", func() {
		client, err := NewClient("https://example.test/nextcloud/?ignored=yes#fragment")
		Expect(err).NotTo(HaveOccurred())
		base := client.BaseURL()
		Expect(base.String()).To(Equal("https://example.test/nextcloud"))
	})

	It("rejects invalid server URLs and nil options", func() {
		_, err := NewClient("://bad")
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
		_, err = NewClient("ftp://example.test")
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
		_, err = NewClient("https:///missing")
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
		_, err = NewClient("https://example.test", nil)
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
	})

	It("sends common headers, authentication, query and JSON payloads", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(Equal("/nextcloud/ocs/v2.php/apps/tables/api/2/echo"))
			Expect(request.URL.Query().Get("format")).To(Equal("json"))
			Expect(request.URL.Query().Get("x")).To(Equal("y"))
			Expect(request.Header.Get("Accept")).To(Equal("application/json"))
			Expect(request.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(request.Header.Get("OCS-APIRequest")).To(Equal("true"))
			username, password, ok := request.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("alice"))
			Expect(password).To(Equal("secret"))
			body, err := io.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(MatchJSON(`{"hello":"world"}`))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":{"value":"done"}}}`)
		}))
		DeferCleanup(server.Close)

		client, err := NewClient(server.URL+"/nextcloud", WithCredentials("alice", "secret"))
		Expect(err).NotTo(HaveOccurred())
		var output struct {
			Value string `json:"value"`
		}
		err = client.doOCS(ctx, client.httpClient, http.MethodPost, "/echo", url.Values{"x": []string{"y"}}, map[string]string{"hello": "world"}, &output)
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Value).To(Equal("done"))
	})

	It("maps HTTP and OCS errors", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/ocs/v2.php/apps/tables/api/2/http":
				writer.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(writer, `{"message":"denied"}`)
			case "/ocs/v2.php/apps/tables/api/2/ocs":
				_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"failure","statuscode":404,"message":"missing"},"data":[]}}`)
			case "/ocs/v2.php/apps/tables/api/2/bad":
				_, _ = io.WriteString(writer, `{bad`)
			case "/ocs/v2.php/apps/tables/api/2/baddata":
				_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":{"x":1}}}`)
			}
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())

		err = client.doOCS(ctx, client.httpClient, http.MethodGet, "/http", nil, nil, nil)
		Expect(IsCode(err, CodeForbidden)).To(BeTrue())
		err = client.doOCS(ctx, client.httpClient, http.MethodGet, "/ocs", nil, nil, nil)
		Expect(IsCode(err, CodeNotFound)).To(BeTrue())
		err = client.doOCS(ctx, client.httpClient, http.MethodGet, "/bad", nil, nil, nil)
		Expect(IsCode(err, CodeInvalidResponse)).To(BeTrue())
		var target []string
		err = client.doOCS(ctx, client.httpClient, http.MethodGet, "/baddata", nil, nil, &target)
		Expect(IsCode(err, CodeInvalidResponse)).To(BeTrue())
	})

	It("maps transport, encode, request, read, and response failures", func(ctx SpecContext) {
		transportErr := errors.New("transport")
		client, err := NewClient("https://example.test", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, transportErr
		})}))
		Expect(err).NotTo(HaveOccurred())
		err = client.do(ctx, client.httpClient, http.MethodGet, "/x", nil, nil, nil)
		Expect(IsCode(err, CodeTransport)).To(BeTrue())

		bad := map[string]any{"ch": make(chan int)}
		err = client.do(ctx, client.httpClient, http.MethodPost, "/x", nil, bad, nil)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())

		badClient := *client
		badClient.baseURL = &url.URL{Scheme: "https", Host: "[::1"}
		err = badClient.do(ctx, client.httpClient, http.MethodGet, "/x", nil, nil, nil)
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())

		client.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(errorReader{}), Header: make(http.Header)}, nil
		})}
		err = client.do(ctx, client.httpClient, http.MethodGet, "/x", nil, nil, nil)
		Expect(IsCode(err, CodeInvalidResponse)).To(BeTrue())
	})

	It("extracts useful response messages", func() {
		Expect(responseMessage([]byte(`{"message":"plain"}`))).To(Equal("plain"))
		Expect(responseMessage([]byte(`{"ocs":{"meta":{"message":"ocs"}}}`))).To(Equal("ocs"))
		Expect(responseMessage([]byte(" raw "))).To(Equal("raw"))
		Expect(responseMessage([]byte(strings.Repeat("x", 241)))).To(HaveLen(240))
		Expect(joinURLPath("/nextcloud", "/a/")).To(Equal("/nextcloud/a/"))
	})
})

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read") }
