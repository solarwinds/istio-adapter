package(default_visibility = ["//mixer/adapter/appoptics:__pkg__"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["logger.go"],
    importpath = "istio.io/istio/mixer/adapter/appoptics/logger",
    deps = [
         "//mixer/adapter/appoptics/config:go_default_library",
        "//mixer/pkg/adapter:go_default_library",
        "//mixer/pkg/pool:go_default_library",
        "//mixer/template/logentry:go_default_library",
        "@com_github_boltdb_bolt//:go_default_library",
        "@com_github_sirupsen_logrus//:go_default_library",
    ],
)