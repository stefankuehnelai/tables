package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func envForServer(server string) func(string) string {
	values := map[string]string{
		envHostname: "test.example",
		envServer:   server,
		envUsername: "alice",
		envPassword: "secret",
	}
	return func(key string) string { return values[key] }
}

var _ = Describe("table commands", func() {
	It("lists tables with structured output", func() {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":[{"id":1,"title":"Customers"}]}}`)
		}))
		DeferCleanup(server.Close)
		deps, _, _, output := testDependencies()
		deps.getenv = envForServer(server.URL)
		command := newRootCommand(deps)
		command.SetArgs([]string{"list", "--json", "id,title"})
		Expect(command.Execute()).To(Succeed())
		Expect(output.String()).To(MatchJSON(`[{"id":1,"title":"Customers"}]`))
	})

	It("creates, gets, updates, and deletes tables", func() {
		requests := make([]string, 0, 4)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests = append(requests, request.Method+" "+request.URL.Path)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":{"id":7,"title":"Customers"}}}`)
		}))
		DeferCleanup(server.Close)
		deps, _, _, _ := testDependencies()
		deps.getenv = envForServer(server.URL)
		for _, args := range [][]string{
			{"create", "--title", "Customers"},
			{"get", "7"},
			{"update", "7", "--title", "Clients", "--archived"},
			{"delete", "7"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed())
		}
		Expect(requests).To(HaveLen(4))
	})
})
