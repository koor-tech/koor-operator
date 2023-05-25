#!/bin/bash
# A workaround until https://github.com/arttor/helmify/issues/67 is solved.
# run this after make
# TODO restore local changes

git diff -U0 charts/koor-operator/values.yaml > patch.diff
sed -z -i 's/\n-\s*# --.*\n/\n/gm' patch.diff
sed -z -i 's/\n@@.*,0.*@@.*//gm' patch.diff
rediff patch.diff | tee patch2.diff
git restore charts/koor-operator/values.yaml
patch -p1 --no-backup-if-mismatch < patch2.diff
rm patch.diff patch2.diff
