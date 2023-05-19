#!/bin/bash
# A workaround until https://github.com/arttor/helmify/issues/67 is solved.
git diff -w -I'^\s*# --' charts/koor-operator/values.yaml
if ((! $?)) ; then
    git checkout charts/koor-operator/values.yaml
fi
