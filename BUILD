load("@bazel_gazelle//:def.bzl", "gazelle")
load("@rules_oci//oci:defs.bzl", "oci_push")
load("//:config.bzl", "registry_api_image", "registry_monitor_image", "registry_user_base_image")

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
    image = "//app/dcr_api:image",
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