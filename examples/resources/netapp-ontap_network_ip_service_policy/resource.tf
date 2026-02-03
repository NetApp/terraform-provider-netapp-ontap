resource "netapp-ontap_network_ip_service_policy" "example" {
    cx_profile_name = "hw-cluster"
    name           = "test_service_policy1"
    svm_name       = "svm1"
    services = [
        "data_nfs",
        "data_cifs",
        "management_ssh"
    ]
    ipspace = "Default"
}
