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
		commandFactory = func() (*cobra.Command, error) { return staticCommand(nil) }
		DeferCleanup(func() { commandFactory = originalFactory })
		Expect(run()).To(Equal(0))
	})

	It("maps public errors to process exit status", func() {
		originalFactory := commandFactory
		DeferCleanup(func() { commandFactory = originalFactory })
		commandFactory = func() (*cobra.Command, error) {
			command := &cobra.Command{Use: "tables", RunE: func(*cobra.Command, []string) error {
				return &tables.Error{ErrorCode: tables.CodeNotFound, Message: "missing"}
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
		commandFactory = func() (*cobra.Command, error) { return staticCommand(nil) }
		called := -1
		exitProcess = func(code int) { called = code }
		DeferCleanup(func() {
			commandFactory = originalFactory
			exitProcess = originalExit
		})
		main()
		Expect(called).To(Equal(0))
	})
})
