package tables

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	It("exposes stable codes and unwraps causes", func() {
		cause := errors.New("boom")
		err := &Error{code: CodeTransport, statusCode: http.StatusBadGateway, message: "transport", cause: cause}
		Expect(err.Error()).To(ContainSubstring("HTTP 502"))
		Expect(err.Code()).To(Equal(CodeTransport))
		Expect(err.HTTPStatus()).To(Equal(http.StatusBadGateway))
		Expect(errors.Is(err, cause)).To(BeTrue())
		Expect(IsCode(err, CodeTransport)).To(BeTrue())
	})

	It("handles nil errors defensively", func() {
		var err *Error
		Expect(err.Error()).To(BeEmpty())
		Expect(err.Code()).To(Equal(CodeUnknown))
		Expect(err.HTTPStatus()).To(BeZero())
		Expect(err.Unwrap()).To(BeNil())
	})

	It("maps HTTP statuses", func() {
		cases := map[int]int{
			http.StatusBadRequest:          CodeInvalidArgument,
			http.StatusUnauthorized:        CodeUnauthorized,
			http.StatusForbidden:           CodeForbidden,
			http.StatusNotFound:            CodeNotFound,
			http.StatusConflict:            CodeConflict,
			http.StatusInternalServerError: CodeServer,
			http.StatusTeapot:              CodeUnknown,
		}
		for status, code := range cases {
			err := errorForStatus(status, "")
			Expect(err.Code()).To(Equal(code))
			Expect(err.Error()).To(ContainSubstring(http.StatusText(status)))
		}
		Expect(NewError(CodeUnknown, "x").Error()).To(Equal("tables: x"))
		Expect(IsCode(errors.New("plain"), CodeUnknown)).To(BeFalse())
	})
})
