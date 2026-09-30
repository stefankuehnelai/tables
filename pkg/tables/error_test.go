package tables_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Errors", func() {
	It("exposes stable error codes and causes", func() {
		By("Arrange")
		cause := errors.New("boom")
		err := &tables.Error{ErrorCode: tables.CodeTransport, Message: "request failed", Cause: cause}

		By("Act")
		message := err.Error()

		By("Assert")
		Expect(err.Code()).To(Equal(tables.CodeTransport))
		Expect(message).To(Equal("request failed: boom"))
		Expect(errors.Is(err, cause)).To(BeTrue())
	})

	It("uses a cause-only message", func() {
		cause := errors.New("boom")
		Expect((&tables.Error{Cause: cause}).Error()).To(Equal("boom"))
	})

	It("handles nil receivers", func() {
		var err *tables.Error
		Expect(err.Code()).To(Equal(tables.CodeOK))
		Expect(err.Error()).To(Equal("<nil>"))
		Expect(err.Unwrap()).To(BeNil())
	})
})
