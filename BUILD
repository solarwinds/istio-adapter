# gazelle:ignore
package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("//mixer/adapter:inventory.bzl", "inventory_library")

inventory_library(
    name = "go_default_library",
    packages = {
        "denier": "istio.io/istio/mixer/adapter/denier",
        "kubernetes": "istio.io/istio/mixer/adapter/kubernetesenv",
        "list": "istio.io/istio/mixer/adapter/list",
        "noop": "istio.io/istio/mixer/adapter/noop",
        "prometheus": "istio.io/istio/mixer/adapter/prometheus",
        "appoptics": "istio.io/istio/mixer/adapter/appoptics",
        "servicecontrol": "istio.io/istio/mixer/adapter/servicecontrol",
        "stackdriver": "istio.io/istio/mixer/adapter/stackdriver",
        "statsd": "istio.io/istio/mixer/adapter/statsd",
        "stdio": "istio.io/istio/mixer/adapter/stdio",
        "memquota": "istio.io/istio/mixer/adapter/memquota",
        "circonus": "istio.io/istio/mixer/adapter/circonus",
    },
    deps = [
        "//mixer/adapter/circonus:go_default_library",
        "//mixer/adapter/denier:go_default_library",
        "//mixer/adapter/kubernetesenv:go_default_library",
        "//mixer/adapter/list:go_default_library",
        "//mixer/adapter/memquota:go_default_library",
        "//mixer/adapter/noop:go_default_library",
        "//mixer/adapter/prometheus:go_default_library",
        "//mixer/adapter/appoptics:go_default_library",
        "//mixer/adapter/servicecontrol:go_default_library",
        "//mixer/adapter/stackdriver:go_default_library",
        "//mixer/adapter/statsd:go_default_library",
        "//mixer/adapter/stdio:go_default_library",
    ],
)
