#!/usr/bin/env bash

set -e

# Start by checking that $TRAVIS_COMMIT_RANGE is a valid range. The range may
# not be reachable if a PR has been forced pushed between the Travis job has
# started and when it clones the git repository.
if [ "$TRAVIS_COMMIT_RANGE" != "" ] && ! git rev-list --quiet $TRAVIS_COMMIT_RANGE; then
  echo "Maybe the PR was forced pushed?"
  exit 1
fi

if ! git diff --name-only $TRAVIS_COMMIT_RANGE | grep -qvE '(\.md)|(\.MD)|(\.png)|(\.pdf)|^(doc/)|^(MAINTAINERS)|^(LICENSE)'; then
  echo "Only doc files were updated, not running the CI."
  exit 0
fi
