#!/bin/bash

set -e

testdirs=$(find . -name 'test*' -type d )

for t in $testdirs; do
    echo "running test in $t"
    kustomize build --enable-alpha-plugins "$t" > "$t"/output.yaml
    diff "$t"/expected.yaml "$t"/output.yaml
done
