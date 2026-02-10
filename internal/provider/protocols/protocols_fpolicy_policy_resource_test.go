package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsFpolicyPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error - non-existent SVM
			{
				Config:      testAccProtocolsFpolicyPolicyResourceConfig("non-existant", "test_policy", []string{"scooby"}),
				ExpectError: regexp.MustCompile("svm non-existant not found"),
			},
			// Create and read testing
			{
				Config: testAccProtocolsFpolicyPolicyResourceConfig("scooby", "acc_fpolicy_policy_test", []string{"scooby"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "name", "acc_fpolicy_policy_test"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "events.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "mandatory", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "passthrough_read", "false"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_fpolicy_policy.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf-acc-policy", "scooby", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "name", "tf-acc-policy"),
					resource.TestCheckResourceAttr("netapp-ontap_fpolicy_policy.example", "svm_name", "scooby"),
				),
			},
		},
	})
}

func testAccProtocolsFpolicyPolicyResourceConfig(svmName string, name string, events []string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}

	eventsStr := "["
	for i, event := range events {
		if i > 0 {
			eventsStr += ", "
		}
		eventsStr += fmt.Sprintf("\"%s\"", event)
	}
	eventsStr += "]"

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

resource "netapp-ontap_fpolicy_policy" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	events = %s
	mandatory = false
	passthrough_read = false
	priority = 5
	scope = {
    include_volumes   = ["scooby_root"]
	}
}`, host, admin, password, svmName, name, eventsStr)
}
