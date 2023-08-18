package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClusterScheduleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test create interval schedule error
			{
				Config:      testAccClusterScheduleResourceIntervalConfig("non-existant", "wrongvalue"),
				ExpectError: regexp.MustCompile("error creating cluster_schedule"),
			},
			// Create intervale and read
			{
				Config: testAccClusterScheduleResourceIntervalConfig("tf-interval-schedule-test", "PT8M30S"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.example", "name", "tf-interval-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.example", "interval", "PT8M30S"),
				),
			},
			// update and read
			{
				Config: testAccClusterScheduleResourceIntervalConfig("tf-interval-schedule-test", "PT8M20S"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.example", "name", "tf-interval-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.example", "interval", "PT8M20S"),
				),
			},
			// Create cron schedule and read
			{
				Config: testAccClusterScheduleCreateResourceCronConfig("tf-cron-schedule-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.cron-example", "name", "tf-cron-schedule-test"),
				),
			},
			// Update cron schedule and read
			{
				Config: testAccClusterScheduleUpdateResourceCronConfig("tf-cron-schedule-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule_resource.cron-example", "name", "tf-cron-schedule-test"),
				),
			},
		},
	})
}

func testAccClusterScheduleResourceIntervalConfig(name string, interval string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  interval = "%s"
}`, host, admin, password, name, interval)
}

func testAccClusterScheduleCreateResourceCronConfig(name string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "cron-example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  cron = {
    minutes = [1, 2, 3, 4]
    hours = [10]
    days = [1, 2]
    months = [6, 7]
    weekdays = [1, 3, 4]
  }
}`, host, admin, password, name)
}

func testAccClusterScheduleUpdateResourceCronConfig(name string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "cron-example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  cron = {
    minutes = [4, 5, 6]
    hours = [2, 3]
    days = [2]
    months = [1, 6, 7]
    weekdays = [3, 4, 5]
  }
}`, host, admin, password, name)
}
