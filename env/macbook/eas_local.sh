#!/usr/bin/env zsh

set -x

export EAS_LOCAL_BUILD_PLUGIN_PATH=/Users/wojciechkozyra/expo/eas-build/bin/eas-cli-local-build-plugin
export EAS_LOCAL_BUILD_WORKINGDIR=/Users/wojciechkozyra/expo/eas-build-workingdir
export EAS_LOCAL_BUILD_SKIP_CLEANUP=1
export EAS_LOCAL_BUILD_ARTIFACTS_DIR=/Users/wojciechkozyra/expo/eas-build-workingdir/results

rm -rf $EAS_LOCAL_BUILD_WORKINGDIR
/Users/wojciechkozyra/expo/eas-cli/bin/run "$@"
