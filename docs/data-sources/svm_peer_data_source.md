---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netapp-ontap_svm_peer_data_source Data Source - terraform-provider-netapp-ontap"
subcategory: "SVM"
description: |-
  Retrieves the configuration of SVM Peer.
---

# Data Source svm peer

Retrieves the configuration of SVM Peer.


## Example Usage

```terraform
data "netapp-ontap_svm_peer_data_source" "example" {
  cx_profile_name = "cluster4"
  svm = {
    name = "test"
  }
  peer = {
    svm = {
      name = "test_peer"
    }
  }
}`
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cx_profile_name` (String) Connection profile name
- `svm` (Attributes) (see [below for nested schema](#nestedatt--svm))
- `peer` (Attributes) (see [below for nested schema](#nestedatt--peer))


### Read-Only

- `id` (String) svm peeer identifier
- `applications` (List of Strings) SVMPeering applications
- `cluster`  (Attributes) (see [below for nested schema](#nestedatt--cluster))
- `state` (String) SVMPeering state

<a id="nestedatt--peer"></a>
### Nested Schema for `peer`

Required:

- `svm` (Attributes) (see [below for nested schema](#nestedatt--svm))

<a id="nestedatt--svm"></a>
### Nested Schema for `svm`

Required:

- `name` (String) name of the SVM.

<a id="nestedatt--cluster"></a>
### Nested Schema for `cluster`

Read-Only:

- `name` (String) name of the Cluster.