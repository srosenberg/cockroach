# Generating the code for `sort_partitioner.eg.go` requires special handling
# because the template lives in a subpackage.
def gen_sort_partitioner_rule(name, target, visibility = ["//visibility:private"]):
    native.genrule(
        name = name,
        srcs = ["//pkg/sql/colexec/colexecbase:distinct_tmpl.go"],
        outs = [target],
        cmd = """\
export COCKROACH_INTERNAL_DISABLE_METAMORPHIC_TESTING=true
export LD_LIBRARY_PATH=/home/srosenberg/go/src/antithesis.com/instrumentation/lib/
$(location :execgen) -template $(SRCS) -fmt=false pkg/sql/colexec/$@ > $@
$(location :goimports) -w $@
""",
        tools = [":execgen", ":goimports"],
        visibility = [":__pkg__", "//pkg/gen:__pkg__"],
    )
