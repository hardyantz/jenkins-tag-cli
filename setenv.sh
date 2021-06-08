#!/bin/bash
echo BRANCH="${1}" > env.properties
echo COMMIT_HASH="${2}" >> env.properties
echo GIT_BRANCH="$1" >> env.properties
git tag -f "${1}" "${2}"
git push -f origin "${1}"