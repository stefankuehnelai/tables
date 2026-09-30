package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("columns", func() {
	It("performs column CRUD", func(ctx SpecContext) {
		table, err := cloud.CreateTable(ctx, cloud.AdminUsername, cloud.AdminPassword, fmt.Sprintf("columns-%d", GinkgoRandomSeed()))
		Expect(err).NotTo(HaveOccurred())
		client, err := cloud.Client(cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func(ctx SpecContext) { _ = client.DeleteTable(ctx, table.ID) })

		By("Create")
		created, err := client.Table(table.ID).Columns().Create(ctx, tables.CreateColumnOptions{Title: "Name", Type: "text", Subtype: "line"})
		Expect(err).NotTo(HaveOccurred())

		By("List")
		listed, err := client.Table(table.ID).Columns().List(ctx, tables.ListColumnsOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ContainElement(HaveField("ID", created.ID)))

		By("Update")
		title := "Display name"
		updated, err := client.Table(table.ID).Columns().Update(ctx, created.ID, tables.UpdateColumnOptions{Title: &title})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.Title).To(Equal(title))

		By("Delete")
		Expect(client.Table(table.ID).Columns().Delete(ctx, created.ID)).To(Succeed())
	})
})
