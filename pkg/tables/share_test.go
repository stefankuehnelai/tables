package tables_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Public shares", func() {
	It("uses public share CRUD without a password", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(ContainSubstring("/api/2/public/aaaaaaaaaaaaaaaa/rows"))
			switch request.Method {
			case http.MethodGet:
				writeOCS(writer, http.StatusOK, []map[string]any{{"id": 5}})
			case http.MethodPost, http.MethodPut:
				writeOCS(writer, http.StatusOK, map[string]any{"id": 5})
			case http.MethodDelete:
				writeOCS(writer, http.StatusOK, map[string]any{})
			}
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := client.Share(tables.Share{Token: tables.ShareToken("aaaaaaaaaaaaaaaa")}).Rows()
		_, err = rows.List(ctx, tables.ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
		row, err := rows.Create(ctx, tables.CreateRowOptions{Values: tables.RowValues{1: "x"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = rows.Update(ctx, row.ID, tables.UpdateRowOptions{Values: tables.RowValues{1: "y"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Delete(ctx, row.ID)).To(Succeed())
	})

	It("isolates password session cookies between concurrent share scopes", func(ctx SpecContext) {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
			if request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/authenticate") {
				token := parts[len(parts)-2]
				Expect(request.ParseForm()).To(Succeed())
				if request.Form.Get("password") != "pw-"+token {
					writer.WriteHeader(http.StatusOK)
					return
				}
				http.SetCookie(writer, &http.Cookie{Name: "share", Value: token, Path: "/"})
				writer.Header().Set("Location", "/")
				writer.WriteHeader(http.StatusFound)
				return
			}
			if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/api/2/public/") {
				token := parts[len(parts)-2]
				cookie, err := request.Cookie("share")
				Expect(err).NotTo(HaveOccurred())
				Expect(cookie.Value).To(Equal(token))
				writeOCS(writer, http.StatusOK, []map[string]any{{"id": len(token)}})
				return
			}
			Fail(fmt.Sprintf("unexpected request %s %s", request.Method, request.URL.Path))
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		first := client.Share(tables.Share{Token: "aaaaaaaaaaaaaaaa", Password: "pw-aaaaaaaaaaaaaaaa"}).Rows()
		second := client.Share(tables.Share{Token: "bbbbbbbbbbbbbbbb", Password: "pw-bbbbbbbbbbbbbbbb"}).Rows()

		By("Act")
		var wait sync.WaitGroup
		wait.Add(2)
		errorsChannel := make(chan error, 2)
		go func() {
			defer wait.Done()
			_, err := first.List(ctx, tables.ListRowsOptions{})
			errorsChannel <- err
		}()
		go func() {
			defer wait.Done()
			_, err := second.List(ctx, tables.ListRowsOptions{})
			errorsChannel <- err
		}()
		wait.Wait()
		close(errorsChannel)

		By("Assert")
		for err := range errorsChannel {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("reports a wrong share password as authentication failure and caches the result", func(ctx SpecContext) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if strings.HasSuffix(request.URL.Path, "/authenticate") {
				calls++
				writer.WriteHeader(http.StatusOK)
				return
			}
			Fail("row request must not run after failed authentication")
		}))
		DeferCleanup(server.Close)
		client, err := tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := client.Share(tables.Share{Token: "aaaaaaaaaaaaaaaa", Password: "wrong"}).Rows()
		_, err = rows.List(ctx, tables.ListRowsOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeAuthenticationFailed))
		_, err = rows.List(ctx, tables.ListRowsOptions{})
		Expect(err).To(HaveOccurred())
		Expect(calls).To(Equal(1))
	})

	It("validates share scopes", func(ctx SpecContext) {
		client, err := tables.NewClient("https://example.net")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(tables.Share{}).Rows().List(ctx, tables.ListRowsOptions{})
		Expect(err).To(HaveOccurred())
		var scope *tables.ShareScope
		Expect(scope.Rows()).NotTo(BeNil())
	})

	It("maps share authentication transport and HTTP failures", func(ctx SpecContext) {
		client, err := tables.NewClient("http://127.0.0.1:1", tables.WithHTTPClient(&http.Client{}))
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(tables.Share{Token: "aaaaaaaaaaaaaaaa", Password: "pw"}).Rows().List(ctx, tables.ListRowsOptions{})
		Expect(err.(tables.TablesError).Code()).To(Equal(tables.CodeTransport))

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusForbidden) }))
		DeferCleanup(server.Close)
		client, err = tables.NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(tables.Share{Token: "aaaaaaaaaaaaaaaa", Password: "pw"}).Rows().List(ctx, tables.ListRowsOptions{})
		var tablesErr tables.TablesError
		Expect(errors.As(err, &tablesErr)).To(BeTrue())
		Expect(tablesErr.Code()).To(Equal(tables.CodeForbidden))
	})

	It("returns invalid argument for a nil share scope authentication path", func() {
		var scope *tables.ShareScope
		// Rows on a nil scope is a zero-valued Rows and is validated before transport.
		_, err := scope.Rows().List(context.Background(), tables.ListRowsOptions{})
		Expect(err).To(HaveOccurred())
	})
})
