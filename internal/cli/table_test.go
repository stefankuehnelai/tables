package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("table commands", func() {
	It("lists tables with structured output", func() {
		By("Arrange")
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			Expect(request.URL.Path).To(Equal("/ocs/v2.php/apps/tables/api/2/tables"))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":[{"id":1,"title":"Customers"}]}}`)
		}))
		DeferCleanup(server.Close)
		deps, _, _, output := testDependencies()
		deps.getenv = envForServer(server.URL)
		command := newRootCommand(deps)
		command.SetArgs([]string{"list", "--json", "id,title"})

		By("Act")
		err := command.Execute()

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(output.String()).To(MatchJSON(`[{"id":1,"title":"Customers"}]`))
	})

	It("creates, gets, updates, and deletes tables", func() {
		By("Arrange")
		requests := make([]string, 0, 4)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests = append(requests, request.Method+" "+request.URL.Path)
			writer.Header().Set("Content-Type", "application/json")
			if request.Method == http.MethodDelete {
				_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":[]}}`)
				return
			}
			_, _ = fmt.Fprint(writer, `{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":{"id":7,"title":"Customers"}}}`)
		}))
		DeferCleanup(server.Close)
		deps, _, _, _ := testDependencies()
		deps.getenv = envForServer(server.URL)

		By("Act and Assert")
		for _, args := range [][]string{
			{"create", "--title", "Customers"},
			{"get", "7"},
			{"update", "7", "--title", "Clients", "--archived"},
			{"delete", "7"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed(), strings.Join(args, " "))
		}
		Expect(requests).To(Equal([]string{
			"POST /ocs/v2.php/apps/tables/api/2/tables",
			"GET /ocs/v2.php/apps/tables/api/2/tables/7",
			"PUT /ocs/v2.php/apps/tables/api/2/tables/7",
			"DELETE /ocs/v2.php/apps/tables/api/2/tables/7",
		}))
	})

	It("rejects invalid table IDs", func() {
		deps, _, _, _ := testDependencies()
		command := newRootCommand(deps)
		command.SetArgs([]string{"get", "nope"})
		Expect(command.Execute()).To(MatchError(ContainSubstring("positive integer")))
	})
})

func envForServer(server string) func(string) string {
	values := map[string]string{
		envHostname: "test.example", envServer: server, envUsername: "alice", envPassword: "secret",
	}
	return func(key string) string { return values[key] }
}
