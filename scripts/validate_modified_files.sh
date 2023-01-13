#!/bin/bash
set -ex

#############
# VARIABLES #
#############
CRD_ERR="changes found by 'make crds'. please run 'make crds' locally and update your PR"

#############
# FUNCTIONS #
#############
function validate(){
  git=$(git status --porcelain)
  for file in $git; do
    if [ -n "$file" ]; then
      echo "$1"
      echo "$git"
      git diff
      exit 1
    fi
  done
}

########
# MAIN #
########
case "$1" in
  crd)
    validate "$CRD_ERR"
  ;;
  *)
    echo $"Usage: $0 {crd}"
    exit 1
esac
