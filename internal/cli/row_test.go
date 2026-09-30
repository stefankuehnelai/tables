package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("row commands", func() {
	It("parses typed and string values", func() {
		values, err := parseRowValues([]string{"1=true", "2=3.5", "3=hello", `4={"x":1}`})
		Expect(err).NotTo(HaveOccurred())
		Expect(values[1]).To(Equal(true))
		Expect(values[2]).To(Equal(3.5))
		Expect(values[3]).To(Equal("hello"))
		Expect(values[4]).To(HaveKeyWithValue("x", float64(1)))
	})

	It("rejects malformed values", func() {
		_, err := parseRowValues([]string{"bad"})
		Expect(err).To(HaveOccurred())
		_, err = parseRowValues([]string{"nope=value"})
		Expect(err).To(HaveOccurred())
	})

	It("runs authenticated row CRUD", func() {
		requests := []string{}
		server := rowServer(&requests)
		DeferCleanup(server.Close)
		deps, _, _, _ := testDependencies()
		deps.getenv = envForServer(server.URL)
		for _, args := range [][]string{
			{"rows", "list", "7", "--limit", "10"},
			{"rows", "create", "7", "--value", "3=Alice"},
			{"rows", "update", "7", "9", "--value", "3=Bob"},
			{"rows", "delete", "7", "9"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed())
		}
		Expect(requests).To(Equal([]string{
			"GET /index.php/apps/tables/api/1/tables/7/rows?limit=10",
			"POST /ocs/v2.php/apps/tables/api/2/tables/7/rows?format=json",
			"PUT /ocs/v2.php/apps/tables/api/2/tables/7/rows/9?format=json",
			"DELETE /ocs/v2.php/apps/tables/api/2/tables/7/rows/9?format=json",
		}))
	})

	It("runs public-share row CRUD without account credentials", func() {
		requests := []string{}
		server := rowServer(&requests)
		DeferCleanup(server.Close)
		deps, _, _, _ := testDependencies()
		deps.getenv = func(key string) string {
			if key == envServer {
				return server.URL
			}
			return ""
		}
		for _, args := range [][]string{
			{"rows", "list", "--share-token", "abcdefghijklmnop"},
			{"rows", "create", "--share-token", "abcdefghijklmnop", "--value", "3=Alice"},
			{"rows", "update", "9", "--share-token", "abcdefghijklmnop", "--value", "3=Bob"},
			{"rows", "delete", "9", "--share-token", "abcdefghijklmnop"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed())
		}
		Expect(requests).To(Equal([]string{
			"GET /ocs/v2.php/apps/tables/api/2/public/abcdefghijklmnop/rows?format=json",
			"POST /ocs/v2.php/apps/tables/api/2/public/abcdefghijklmnop/rows?format=json",
			"PUT /ocs/v2.php/apps/tables/api/2/public/abcdefghijklmnop/rows/9?format=json",
			"DELETE /ocs/v2.php/apps/tables/api/2/public/abcdefghijklmnop/rows/9?format=json",
		}))
	})

	It("rejects ambiguous row targets", func() {
		deps, _, _, _ := testDependencies()
		deps.getenv = func(key string) string {
			if key == envServer {
				return "https://example.net"
			}
			return ""
		}
		command := newRootCommand(deps)
		command.SetArgs([]string{"rows", "list", "7", "--share-token", "abcdefghijklmnop"})
		Expect(command.Execute()).To(MatchError(ContainSubstring("do not combine")))
	})
})

func rowServer(requests *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		if request.URL.RawQuery != "" {
			path += "?" + request.URL.RawQuery
		}
		*requests = append(*requests, request.Method+" "+path)
		writer.Header().Set("Content-Type", "application/json")
		data := `{"id":9,"tableId":7,"data":[{"columnId":3,"value":"Alice"}]}`
		if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/index.php/apps/tables/api/1/") {
			_, _ = fmt.Fprint(writer, `[`+data+`]`)
			return
		}
		if request.Method == http.MethodDelete {
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":`+data+`}}`)
			return
		}
		if request.Method == http.MethodGet {
			data = `[` + data + `]`
		}
		_, _ = fmt.Fprintf(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":%s}}`, data)
	}))
}
