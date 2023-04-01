#!/usr/bin/env zsh

export EAS_LOCAL_BUILD_PLUGIN_PATH=/home/wojtek/expo/eas-build/bin/eas-cli-local-build-plugin
export EAS_LOCAL_BUILD_WORKINGDIR=/home/wojtek/expo/eas-build-workingdir
export EAS_LOCAL_BUILD_SKIP_CLEANUP=1
export EAS_LOCAL_BUILD_ARTIFACTS_DIR=/home/wojtek/expo/eas-build-workingdir/results

rm -rf $EAS_LOCAL_BUILD_WORKINGDIR
/home/wojtek/expo/eas-cli/packages/eas-cli/bin/run "$@"
