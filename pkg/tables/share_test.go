package tables

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Public shares", func() {
	It("authenticates a password-protected share once and reuses its isolated cookie", func(ctx SpecContext) {
		var logins atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/index.php/apps/tables/s/token/authenticate" {
				logins.Add(1)
				Expect(request.FormValue("password")).To(Equal("secret"))
				http.SetCookie(writer, &http.Cookie{Name: "share", Value: "token", Path: "/"})
				writer.Header().Set("Location", "/share")
				writer.WriteHeader(http.StatusFound)
				return
			}
			cookie, err := request.Cookie("share")
			Expect(err).NotTo(HaveOccurred())
			Expect(cookie.Value).To(Equal("token"))
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":[]}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := client.Share(Share{Token: ShareToken("token"), Password: SharePassword("secret")}).Rows()

		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				_, callErr := rows.List(ctx, ListRowsOptions{})
				Expect(callErr).NotTo(HaveOccurred())
			}()
		}
		wg.Wait()
		Expect(logins.Load()).To(Equal(int32(1)))
	})

	It("does not leak cookies between share scopes", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodPost {
				token := request.URL.Path[len("/index.php/apps/tables/s/") : len(request.URL.Path)-len("/authenticate")]
				http.SetCookie(writer, &http.Cookie{Name: "share", Value: token, Path: "/"})
				writer.Header().Set("Location", "/")
				writer.WriteHeader(http.StatusFound)
				return
			}
			cookie, err := request.Cookie("share")
			Expect(err).NotTo(HaveOccurred())
			Expect(request.URL.Path).To(ContainSubstring("/public/" + cookie.Value + "/rows"))
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":[]}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		one := client.Share(Share{Token: "one", Password: "pw"}).Rows()
		two := client.Share(Share{Token: "two", Password: "pw"}).Rows()
		_, err = one.List(ctx, ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
		_, err = two.List(ctx, ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("supports unprotected shares and rejects invalid authentication", func(ctx SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodPost {
				writer.WriteHeader(http.StatusOK)
				return
			}
			_, _ = io.WriteString(writer, `{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":[]}}`)
		}))
		DeferCleanup(server.Close)
		client, err := NewClient(server.URL)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(Share{Token: "public"}).Rows().List(ctx, ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Share(Share{}).Rows().List(ctx, ListRowsOptions{})
		Expect(IsCode(err, CodeInvalidArgument)).To(BeTrue())
		_, err = client.Share(Share{Token: "bad", Password: "wrong"}).Rows().List(ctx, ListRowsOptions{})
		Expect(IsCode(err, CodeAuthenticationFailed)).To(BeTrue())
	})
})
