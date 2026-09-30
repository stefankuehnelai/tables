package integration_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/internal/testenv/nextcloud"
	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("public shares", func() {
	It("supports password-protected full row CRUD", func(ctx SpecContext) {
		table, err := cloud.CreateTable(ctx, cloud.AdminUsername, cloud.AdminPassword, fmt.Sprintf("share-%d", GinkgoRandomSeed()))
		Expect(err).NotTo(HaveOccurred())
		column, err := cloud.CreateColumn(ctx, cloud.AdminUsername, cloud.AdminPassword, table.ID, tables.CreateColumnOptions{Title: "Name", Type: "text", Subtype: "line"})
		Expect(err).NotTo(HaveOccurred())
		owner, err := cloud.Client(cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func(ctx SpecContext) { _ = owner.DeleteTable(ctx, table.ID) })
		token, err := cloud.CreateShare(ctx, cloud.AdminUsername, cloud.AdminPassword, table.ID, "correct-horse", nextcloud.SharePermissions{Read: true, Create: true, Update: true, Delete: true})
		Expect(err).NotTo(HaveOccurred())

		public, err := tables.NewClient(cloud.URL)
		Expect(err).NotTo(HaveOccurred())
		rows := public.Share(tables.Share{Token: tables.ShareToken(token), Password: tables.SharePassword("correct-horse")}).Rows()
		created, err := rows.Create(ctx, tables.CreateRowOptions{Values: tables.RowValues{column.ID: "Alice"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = rows.List(ctx, tables.ListRowsOptions{})
		Expect(err).NotTo(HaveOccurred())
		_, err = rows.Update(ctx, created.ID, tables.UpdateRowOptions{Values: tables.RowValues{column.ID: "Bob"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Delete(ctx, created.ID)).To(Succeed())
	})

	It("rejects an incorrect share password without leaking authentication", func(ctx SpecContext) {
		table, err := cloud.CreateTable(ctx, cloud.AdminUsername, cloud.AdminPassword, fmt.Sprintf("password-%d", GinkgoRandomSeed()))
		Expect(err).NotTo(HaveOccurred())
		owner, err := cloud.Client(cloud.AdminUsername, cloud.AdminPassword)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func(ctx SpecContext) { _ = owner.DeleteTable(ctx, table.ID) })
		token, err := cloud.CreateShare(ctx, cloud.AdminUsername, cloud.AdminPassword, table.ID, "right-password", nextcloud.SharePermissions{Read: true})
		Expect(err).NotTo(HaveOccurred())
		public, err := tables.NewClient(cloud.URL)
		Expect(err).NotTo(HaveOccurred())

		_, err = public.Share(tables.Share{Token: tables.ShareToken(token), Password: "wrong-password"}).Rows().List(ctx, tables.ListRowsOptions{})
		var tablesError tables.TablesError
		Expect(errors.As(err, &tablesError)).To(BeTrue())
		Expect(tablesError.Code()).To(Equal(tables.CodeAuthenticationFailed))
	})
})
