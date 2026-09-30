package integration_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("authentication", func() {
	It("accepts valid password and app-password credentials", func(ctx SpecContext) {
		By("Arrange")
		client, err := tables.NewClient(cloud.URL, tables.WithCredentials(cloud.AdminUsername, cloud.AdminPassword))
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})

		By("Assert")
		Expect(err).NotTo(HaveOccurred())

		By("Create an app password")
		appPassword, err := cloud.CreateAppPassword(ctx, cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		appClient, err := tables.NewClient(cloud.URL, tables.WithAppPassword(cloud.AdminUsername, appPassword))
		Expect(err).NotTo(HaveOccurred())
		_, err = appClient.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("maps invalid credentials to an authentication error", func(ctx SpecContext) {
		By("Arrange")
		client, err := tables.NewClient(cloud.URL, tables.WithCredentials(cloud.AdminUsername, "wrong-password"))
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		_, err = client.ListTables(ctx, tables.ListTablesOptions{})

		By("Assert")
		var tablesError tables.TablesError
		Expect(errors.As(err, &tablesError)).To(BeTrue())
		Expect(tablesError.Code()).To(Or(Equal(tables.CodeAuthenticationFailed), Equal(tables.CodeUnauthorized)))
	})
})
