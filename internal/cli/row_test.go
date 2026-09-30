package cli

import (
	"encoding/json"
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

	It("runs authenticated and public-share row CRUD", func() {
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
			{"rows", "list", "--share-token", "abcdefghijklmnop"},
			{"rows", "create", "--share-token", "abcdefghijklmnop", "--value", "3=Alice"},
			{"rows", "update", "9", "--share-token", "abcdefghijklmnop", "--value", "3=Bob"},
			{"rows", "delete", "9", "--share-token", "abcdefghijklmnop"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed())
		}
		Expect(requests).To(HaveLen(8))
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
		data := map[string]any{"id": 9, "tableId": 7, "data": []map[string]any{{"columnId": 3, "value": "Alice"}}}
		if strings.HasPrefix(request.URL.Path, "/index.php/") {
			_ = json.NewEncoder(writer).Encode([]any{data})
			return
		}
		var response any = data
		if request.Method == http.MethodGet {
			response = []any{data}
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{"ocs": map[string]any{"meta": map[string]any{"status": "ok", "statuscode": 100, "message": "OK"}, "data": response}})
	}))
}
