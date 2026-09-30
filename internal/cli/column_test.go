package cli

import (
	"encoding/json"
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
			column := map[string]any{"id": 3, "tableId": 7, "title": "Name", "type": "text", "subtype": "line"}
			switch request.Method {
			case http.MethodGet:
				_ = json.NewEncoder(writer).Encode([]any{column})
			case http.MethodDelete:
				_ = json.NewEncoder(writer).Encode(column)
			default:
				_ = json.NewEncoder(writer).Encode(column)
			}
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
		Expect(requests).To(HaveLen(4))
	})
})
