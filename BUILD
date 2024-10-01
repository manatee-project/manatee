load("@bazel_gazelle//:def.bzl", "gazelle")
load("@rules_oci//oci:defs.bzl", "oci_push")
load("//:config.bzl", 
    "registry_api_image", 
    "registry_monitor_image", 
    "registry_user_base_image",
    "gcp_project_id",
    "gcp_project_number",
    "deploy_env",
    "gcp_region",
    "gcp_zone"
)


# Generate app config file
# TODO: this is temporary solution. Please remove in the future.
sh_binary(
    name = "generate_app_config_sh",
    srcs = [":generate_app_config_file.sh"],
    args = [
       gcp_project_id,
       gcp_project_number,
       deploy_env,
       gcp_region,
       gcp_zone,
    ]
)
genrule(
    name = "generate_app_config",
    srcs = [],
    outs = ["config.yaml"],
    cmd = "$(location :generate_app_config_sh) {} {} {} {} {} > $(OUTS)".format(
        gcp_project_id, 
        gcp_project_number, 
        deploy_env, 
        gcp_region, 
        gcp_zone
    ),
    tools = [":generate_app_config_sh"],
    visibility = ["//visibility:public"],
)

# gazelle:prefix github.com/manatee-project/manatee
gazelle(name = "gazelle")

# push images
oci_push(
    name = "push_dcr_api_image",
    image = "//app/dcr_api:image",
    repository = registry_api_image,
    remote_tags = ["latest"],
)
oci_push(
    name = "push_dcr_monitor_image",
    image = "//app/dcr_monitor:image",
    repository = registry_monitor_image,
    remote_tags = ["latest"],
)

# push user images
oci_push(
    name = "push_dcr_tee_image",
    image = "//app/dcr_tee:image",
    repository = registry_user_base_image,
    remote_tags = ["latest"],
)