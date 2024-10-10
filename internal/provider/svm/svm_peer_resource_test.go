package svm_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSvmPeerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test cluster peer non existant to do svm peer
			{
				Config:      testAccSvmPeerResourceConfig("testme", "testme2", "abcd", "snapmirror"),
				ExpectError: regexp.MustCompile("9895941"),
			},
			// Testing in VSIM is failing to peer
			// // Create svm peer and read
			// {
			// 	Config: testAccSvmPeersResourceConfig("acc_test_peer2", "acc_test2", "swenjuncluster-1", "snapmirror"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_svm_peers.example", "svm.name", "acc_test_peer2"),
			// 	),
			// },
			// // Update applications
			// {
			// 	Config: testAccSvmPeersResourceConfig("acc_test_peer2", "acc_test2", "swenjuncluster-1", "flexcache"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_svm_peers.example", "applications.0", "flexcache"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_svm_peers.example", "svm.name", "acc_test_peer2"),
			// 	),
			// },
			// Import and read
			{
				ResourceName:  "netapp-ontap_svm_peer.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "terraform", "tf_peer", "swenjuncluster-1", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_peer.example", "svm.name", "snapmirror_dest_dp"),
				),
			},
		},
	})
}
func testAccSvmPeerResourceConfig(svm, peerSvm, peerCluster, applications string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	password2 := os.Getenv("TF_ACC_NETAPP_PASS2")
	host2 := os.Getenv("TF_ACC_NETAPP_HOST2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST2, TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, TF_ACC_NETAPP_PASS2 and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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

resource "netapp-ontap_svm_peer" "example" {
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
}`, host, admin, password2, host2, admin, password2, svm, peerSvm, peerCluster, applications)
}
