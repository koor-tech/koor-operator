#!/bin/bash
# A workaround until https://github.com/operator-framework/operator-sdk/issues/6285 is solved.
git diff --quiet -I'^    createdAt: ' bundle
if ((! $?)) ; then
    git checkout bundle
fi
