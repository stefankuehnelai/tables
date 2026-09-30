package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("rows", func() {
	It("uses the same scoped row API for authenticated CRUD", func(ctx SpecContext) {
		table, err := cloud.CreateTable(ctx, cloud.AdminUsername, cloud.AdminPassword, fmt.Sprintf("rows-%d", GinkgoRandomSeed()))
		Expect(err).NotTo(HaveOccurred())
		column, err := cloud.CreateColumn(ctx, cloud.AdminUsername, cloud.AdminPassword, table.ID, tables.CreateColumnOptions{Title: "Name", Type: "text", Subtype: "line"})
		Expect(err).NotTo(HaveOccurred())
		client, err := cloud.Client(cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func(ctx SpecContext) { _ = client.DeleteTable(ctx, table.ID) })
		rows := client.Table(table.ID).Rows()

		created, err := rows.Create(ctx, tables.CreateRowOptions{Values: tables.RowValues{column.ID: "Alice"}})
		Expect(err).NotTo(HaveOccurred())
		listed, err := rows.List(ctx, tables.ListRowsOptions{Limit: 100})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ContainElement(HaveField("ID", created.ID)))
		updated, err := rows.Update(ctx, created.ID, tables.UpdateRowOptions{Values: tables.RowValues{column.ID: "Bob"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.ID).To(Equal(created.ID))
		Expect(rows.Delete(ctx, created.ID)).To(Succeed())
	})
})
