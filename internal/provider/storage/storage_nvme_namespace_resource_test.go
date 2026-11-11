package storage_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
)

func TestAccStorageNvmeNamespaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create storage nvme namespace and read
			{
				Config: testAccStorageNvmeNamespaceResourceConfig("/vol/tf_acc_volume/ns_tf_acc", "tf_acc_svm", "linux", 10, "kb", 4096),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "name", "/vol/tf_acc_volume/ns_tf_acc"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "ostype", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size", "10"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size_unit", "kb"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.block_size", "4096"),
				),
			},
			// Update size
			{
				Config: testAccStorageNvmeNamespaceResourceConfig("/vol/tf_acc_volume/ns_tf_acc", "tf_acc_svm", "linux", 12, "kb", 4096),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "name", "/vol/tf_acc_volume/ns_tf_acc"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size", "12"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size_unit", "kb"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.block_size", "4096"),
				),
			},
			// Update size_unit
			{
				Config: testAccStorageNvmeNamespaceResourceConfig("/vol/tf_acc_volume/ns_tf_acc", "tf_acc_svm", "linux", 1, "mb", 4096),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "name", "/vol/tf_acc_volume/ns_tf_acc"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_nvme_namespace.example", "space.size_unit", "mb"),
				),
			},
		},
	})
}

func testAccStorageNvmeNamespaceResourceConfig(name string, svmName string, osType string, size int64, sizeUnit string, blockSize int64) string {
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

resource "netapp-ontap_nvme_namespace" "example" {
	cx_profile_name = "cluster4"
	name = "%s"
  svm_name = "%s"
	ostype = "%s"
  space = {
		size = "%d"
		size_unit = "%s"
		block_size = "%d"
  }
}

`, host, admin, password, name, svmName, osType, size, sizeUnit, blockSize)
}
