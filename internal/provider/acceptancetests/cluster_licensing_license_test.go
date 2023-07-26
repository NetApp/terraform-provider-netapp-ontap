package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestLicensingLicenseResouce(t *testing.T) {
	testLicense := os.Getenv("TF_ACC_NETAPP_LICENSE")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLicensingLicenseResourceConfig("testme"),
				ExpectError: regexp.MustCompile("1115159"),
			},
			{
				Config: testAccLicensingLicenseResourceConfig(testLicense),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_licensing_license_resource.cluster_licensing_license", "name", "insight_balance")),
			},
		},
	})
}

func testAccLicensingLicenseResourceConfig(key string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cluster_licensing_license_resource" "cluster_licensing_license" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  keys = ["%s"]
}
`, host, admin, password, key)
}
