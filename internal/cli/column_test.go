package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("column commands", func() {
	It("runs column CRUD through the library", func() {
		requests := []string{}
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests = append(requests, request.Method+" "+request.URL.Path)
			writer.Header().Set("Content-Type", "application/json")
			if request.Method == http.MethodGet {
				_, _ = fmt.Fprint(writer, `[{"id":3,"tableId":7,"title":"Name","type":"text","subtype":"line"}]`)
				return
			}
			_, _ = fmt.Fprint(writer, `{"id":3,"tableId":7,"title":"Name","type":"text","subtype":"line"}`)
		}))
		DeferCleanup(server.Close)
		deps, _, _, _ := testDependencies()
		deps.getenv = envForServer(server.URL)
		for _, args := range [][]string{
			{"columns", "list", "7"},
			{"columns", "create", "7", "--title", "Name", "--type", "text", "--subtype", "line"},
			{"columns", "update", "7", "3", "--title", "Full name"},
			{"columns", "delete", "7", "3"},
		} {
			command := newRootCommand(deps)
			command.SetArgs(args)
			Expect(command.Execute()).To(Succeed())
		}
		Expect(requests).To(Equal([]string{
			"GET /index.php/apps/tables/api/1/tables/7/columns",
			"POST /index.php/apps/tables/api/1/columns",
			"PUT /index.php/apps/tables/api/1/columns/3",
			"DELETE /index.php/apps/tables/api/1/columns/3",
		}))
	})
})
