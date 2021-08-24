#!/bin/bash

set -e

testdirs=$(find . -name 'test*' -type d )

for t in $testdirs; do
    echo "running test in $t"
    kustomize build --enable_alpha_plugins "$t" > "$t"/output.yaml
    diff "$t"/expected.yaml "$t"/output.yaml
done
