load("@bazel_gazelle//:def.bzl", "gazelle")
load("@rules_multirun//:defs.bzl", "multirun")
load("@rules_oci//oci:defs.bzl", "oci_push")
load("//:env.bzl", "env", "project_id", "region", "zone")

# NOTE: this is temporary solution for making config file for the app.
# Ideally, the config should be provisioned at the deployment, not at the build.
# Right now, we just generate one so that we can test the app.
genrule(
    name = "generate_app_config",
    outs = ["config.yaml"],
    cmd = """
echo 'CloudProvider:
  GCP:
    Project: "{project_id}"
    HubBucket: "dcr-{env}-hub"
    Zone: "{zone}"
    Region: "{region}"
    Debug: false
    Env: {env}' > $@
    """.format(
        env = env,
        project_id = project_id,
        region = region,
        zone = zone,
    ),
    visibility = ["//visibility:public"],
)

# gazelle:prefix github.com/manatee-project/manatee
gazelle(name = "gazelle")

REPOS = {
    "dcr_api": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/data-clean-room-api".format(project_id, env),
    "dcr_monitor": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/data-clean-room-monitor".format(project_id, env),
    "jupyterlab_manatee": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/scipy-notebook-with-dcr".format(project_id, env),
    "dcr_tee": "us-docker.pkg.dev/{}/dcr-{}-user-images/data-clean-room-base".format(project_id, env),
}

[
    genrule(
        name = "{}_repo".format(k),
        outs = ["{}_repo.txt".format(k)],
        cmd = "echo '{}' | envsubst > $@".format(v),
    )
    for (k, v) in REPOS.items()
]

[
    oci_push(
        name = "push_{}_image".format(k),
        image = "//app/{}:image".format(k),
        remote_tags = ["latest"],
        repository_file = ":{}_repo".format(k),
    )
    for k in REPOS.keys()
]

multirun(
    name = "push_all_images",
    commands = [
        "push_{}_image".format(k)
        for k in REPOS.keys()
    ],
    jobs = 0,
)
