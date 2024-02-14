package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSvmPeersResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test cluster peer non existant to do svm peer
			{
				Config:      testAccSvmPeersResourceConfig("testme", "testme2", "abcd", "snapmirror"),
				ExpectError: regexp.MustCompile("9895941"),
			},
			// Create svm peer and read
			{
				Config: testAccSvmPeersResourceConfig("acc_test_peer2", "acc_test2", "swenjuncluster-1", "snapmirror"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_peers_resource.example", "svm.name", "acc_test_peer2"),
				),
			},
			// Update applications
			{
				Config: testAccSvmPeersResourceConfig("acc_test_peer2", "acc_test2", "swenjuncluster-1", "flexcache"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_peers_resource.example", "applications.0", "flexcache"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_peers_resource.example", "svm.name", "acc_test_peer2"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_svm_peers_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "acc_test_peer", "acc_test", "swenjuncluster-1", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_peers_resource.example", "svm.name", "acc_test_peer"),
				),
			},
		},
	})
}
func testAccSvmPeersResourceConfig(svm, peer_svm, peer_cluster, applications string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST3")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	host2 := os.Getenv("TF_ACC_NETAPP_HOST2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST2, TF_ACC_NETAPP_HOST3, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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
	{
		name = "cluster3"
		hostname = "%s"
		username = "%s"
		password = "%s"
		validate_certs = false
	},
  ]
}

resource "netapp-ontap_svm_peers_resource" "example" {
  cx_profile_name = "cluster4"
  svm = {
    name = "%s"
  }
  peer = {
    svm = {
      name = "%s"
    }
    cluster = {
      name = "%s"
    }
    peer_cx_profile_name = "cluster3"
  }
  applications = ["%s"]
}`, host, admin, password, host2, admin, password, svm, peer_svm, peer_cluster, applications)
}
