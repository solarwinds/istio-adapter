package(default_visibility = ["//mixer/adapter/appoptics:__pkg__"])

load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "batching.go",
    ],
    importpath = "istio.io/istio/mixer/adapter/appoptics/promadapter",
    visibility = ["//mixer/adapter/appoptics:__pkg__"],
    deps = [
        "//mixer/pkg/adapter:go_default_library",
        "//mixer/adapter/appoptics/appoptics:go_default_library",
    ],
)
