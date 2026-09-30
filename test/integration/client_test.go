package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("client", func() {
	It("uses the real Tables OCS endpoint", func(ctx SpecContext) {
		client, err := tables.NewClient(cloud.URL, tables.WithCredentials(cloud.AdminUsername, cloud.AdminPassword))
		Expect(err).NotTo(HaveOccurred())
		result, err := client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
	})
})
