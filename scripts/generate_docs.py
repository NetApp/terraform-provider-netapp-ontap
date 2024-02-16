import subprocess
import os
import sys

# TO find the correct Catagory, check REST API to see what main header this API lives under

CATAGORYS = {
    'application': [],
    'cloud': [],
    'cluster': [
        "cluster_data_source.md",
        "cluster_schedule_data_source.md",
        "cluster_schedule_resource.md",
        "cluster_licensing_license_resource.md"],
    'nas': [
        "protocols_cifs_local_group_data_source.md",
        "protocols_cifs_local_group_resource.md",
        "protocols_cifs_local_group_member_data_source.md",
        "protocols_cifs_local_group_member_resource.md",
        "protocols_cifs_local_user_data_source.md",
        "protocols_cifs_local_user_resource.md",
        "protocols_cifs_service_data_source.md",
        "protocols_cifs_user_group_privilege_data_source.md",
        "protocols_cifs_user_group_privilege_resource.md",
        "protocols_nfs_service_data_source.md",
        "protocols_nfs_service_resource.md",
        "protocols_nfs_export_policy_resource.md",
        "protocols_nfs_export_policy_rule_data_source.md",
        "protocols_nfs_export_policy_rule_resource.md"],
    'name-services': [
        "name_services_dns_data_source.md",
        "name_services_dns_resource.md"
    ],

    'ndmp': [],
    'networking': [
        "networking_ip_interfaces_data_source.md",
        "networking_ip_interface_data_source.md",
        "networking_ip_interface_resource.md",
        "networking_ip_route_data_source.md",
        "networking_ip_route_resource.md"],
    'nvme': [],
    'object-store': [],
    'san': [],
    'security': [],
    'snaplock': [],
    'snapmirror': ["snapmirror_policy_resource.md"],
    'storage': [
        "storage_aggregate_resource.md",
        "storage_snapshot_policy_resource.md",
        "storage_volume_snapshot_data_source.md",
        "storage_volume_resource.md",
        "storage_volume_data_source.md",
        "storage_volume_snapshot_resource.md"],
    'support': [],
    'svm': ["svm_resource.md"],
}


def main():
    print("===== Generating docs =====")
    generate_doc()
    remove_example()
    print("===== Adding Catagories =====")
    add_catagories()
    print("===== Validate =====")
    validate()
    print("===== Errors =====")
    issue = warn_missing_catagory(["docs/data-sources/", "docs/resources/"])
    if issue:
        sys.exit(1)


def generate_doc():
    cmd_str = "go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate"
    subprocess.run(cmd_str, shell=True)


def validate():
    cmd_str = "go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate"
    subprocess.run(cmd_str, shell=True)


def remove_example():
    files = ['docs/data-sources/example.md', 'docs/resources/example.md']
    for file_path in files:
        if os.path.exists(file_path):
            os.remove(file_path)


def add_catagories():
    for catagory in CATAGORYS:
        for page in CATAGORYS[catagory]:
            if 'data_source' in page:
                update_datasource(page, catagory)
            if 'resource' in page:
                update_resouces(page, catagory)


def update_datasource(page, catagory):
    path = "docs/data-sources/" + page
    update_md_file(path, catagory)


def update_resouces(page, catagory):
    path = "docs/resources/" + page
    update_md_file(path, catagory)


def update_md_file(path, catagory):
    print("Updating %s" % path)
    try:
        with open(path) as f:
            lines = f.readlines()
        for i, line in enumerate(lines):
            if line.startswith('subcategory: "'):
                split_line = line.split('subcategory: "')
                new_line = split_line[0] + 'subcategory: "' + catagory + split_line[1]
                lines[i] = new_line
                break
        with open(path, 'w') as f:
            f.writelines(lines)
    except:
        return

def warn_missing_catagory(directory_paths):
    issue = False
    for directory_path in directory_paths:
        full_path = os.path.join(os.getcwd(), directory_path)
        for filename in os.listdir(full_path):
            if os.path.isfile(os.path.join(full_path, filename)):
                with open(os.path.join(full_path, filename), 'r') as f:
                    file_content = f.read()
                    if 'subcategory: ""' in file_content:
                        print('%s is missing a catagory' % filename)
                        issue = True
    return issue


if __name__ == "__main__":
    main()