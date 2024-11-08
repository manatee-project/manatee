load("@bazel_gazelle//:def.bzl", "gazelle")
load("@rules_multirun//:defs.bzl", "multirun")
load("@rules_oci//oci:defs.bzl", "oci_push")
load("//:env.bzl", "env", "project_id", "region", "zone")

# Generate app config file
# TODO: this is temporary solution. Please remove in the future.
sh_binary(
    name = "generate_app_config_sh",
    srcs = [":generate_app_config_file.sh"],
    args = [
        project_id,
        env,
        region,
        zone,
    ],
)

genrule(
    name = "generate_app_config",
    srcs = [],
    outs = ["config.yaml"],
    cmd = "$(location :generate_app_config_sh) {} {} {} {} > $(OUTS)".format(
        project_id,
        env,
        region,
        zone,
    ),
    tools = [":generate_app_config_sh"],
    visibility = ["//visibility:public"],
)

# gazelle:prefix github.com/manatee-project/manatee
gazelle(name = "gazelle")

REPOS = {
    "dcr_api": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/data-clean-room-api".format(project_id,env),
    "dcr_monitor":  "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/data-clean-room-monitor".format(project_id,env),
    "jupyterlab_manatee": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/scipy-notebook-with-dcr".format(project_id,env),
    "dcr_tee": "us-docker.pkg.dev/{}/dcr-{}-user-images/data-clean-room-base".format(project_id,env),
}

[
    genrule(
        name = "{}_repo".format(k),
        outs = ["{}_repo.txt".format(k)],
        cmd = "echo '{}' | envsubst > $@".format(v)
    )
    for (k, v) in REPOS.items()
]
[
    oci_push(
        name = "push_{}_image".format(k),
        image = "//app/{}:image".format(k),
        remote_tags = ["latest"],
        repository_file = ":{}_repo".format(k)
    )
    for (k, v) in REPOS.items()
]

multirun(
    name = "push_all_images",
    commands = [
        "push_{}_image".format(k)
        for (k, _) in REPOS.items()
    ],
    jobs = 0,
)
