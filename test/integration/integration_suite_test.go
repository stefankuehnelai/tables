package integration_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/internal/testenv/nextcloud"
)

var cloud *nextcloud.Nextcloud

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	options := []nextcloud.Option{}
	if image := os.Getenv("NEXTCLOUD_TABLES_TEST_NEXTCLOUD_IMAGE"); image != "" {
		options = append(options, nextcloud.WithImage(image))
	}
	if version := os.Getenv("NEXTCLOUD_TABLES_TEST_TABLES_VERSION"); version != "" {
		options = append(options, nextcloud.WithTablesVersion(version))
	}
	var err error
	cloud, err = nextcloud.New(ctx, options...)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(cloud.Close)
})
