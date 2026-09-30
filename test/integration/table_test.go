package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("tables", func() {
	It("creates, reads, updates, lists, and deletes a table", func(ctx SpecContext) {
		client, err := cloud.Client(cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		title := fmt.Sprintf("integration-%d", GinkgoRandomSeed())

		By("Create")
		created, err := client.CreateTable(ctx, tables.CreateTableOptions{Title: title, Description: "integration"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.ID).To(BeNumerically(">", 0))

		By("Read")
		fetched, err := client.GetTable(ctx, created.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fetched.Title).To(Equal(title))

		By("List")
		listed, err := client.ListTables(ctx, tables.ListTablesOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ContainElement(HaveField("ID", created.ID)))

		By("Update")
		updatedTitle := title + "-updated"
		updated, err := client.UpdateTable(ctx, created.ID, tables.UpdateTableOptions{Title: &updatedTitle})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.Title).To(Equal(updatedTitle))

		By("Delete")
		Expect(client.DeleteTable(ctx, created.ID)).To(Succeed())
		_, err = client.GetTable(ctx, created.ID)
		Expect(err).To(HaveOccurred())
	})
})
