package main

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("tables executable", func() {
	It("returns zero for success", func() {
		originalFactory := commandFactory
		DeferCleanup(func() { commandFactory = originalFactory })
		commandFactory = func() (*cobra.Command, error) { return &cobra.Command{Use: "tables"}, nil }
		Expect(run()).To(Equal(tables.CodeOK))
	})

	It("maps public errors to process exit status", func() {
		originalFactory := commandFactory
		DeferCleanup(func() { commandFactory = originalFactory })
		commandFactory = func() (*cobra.Command, error) {
			command := &cobra.Command{Use: "tables", RunE: func(*cobra.Command, []string) error {
				return tables.NewError(tables.CodeNotFound, "missing")
			}}
			command.SilenceErrors = true
			command.SilenceUsage = true
			return command, nil
		}
		Expect(run()).To(Equal(tables.CodeNotFound))
	})

	It("maps factory and unknown errors", func() {
		originalFactory := commandFactory
		DeferCleanup(func() { commandFactory = originalFactory })
		commandFactory = func() (*cobra.Command, error) { return nil, errors.New("factory") }
		Expect(run()).To(Equal(tables.CodeUnknown))
		Expect(exitCode(errors.New("plain"))).To(Equal(tables.CodeUnknown))
	})

	It("calls the process exit hook from main", func() {
		originalFactory := commandFactory
		originalExit := exitProcess
		DeferCleanup(func() {
			commandFactory = originalFactory
			exitProcess = originalExit
		})
		commandFactory = func() (*cobra.Command, error) { return &cobra.Command{Use: "tables"}, nil }
		called := -1
		exitProcess = func(code int) { called = code }
		main()
		Expect(called).To(Equal(tables.CodeOK))
	})
})
