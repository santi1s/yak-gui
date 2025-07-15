# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.0](https://github.com/santi1s/yak-gui/compare/v1.7.0...v1.8.0) (2025-07-15)


### Features

* add comprehensive backend testing infrastructure ([1804e0e](https://github.com/santi1s/yak-gui/commit/1804e0e9492412d837e169a50a53d478b627f19a))
* add environment configuration and rollout drill-down features ([9f26b6b](https://github.com/santi1s/yak-gui/commit/9f26b6b85641ceb88409dcba91221a0bb7327364))
* add environment profiles, version display, and ArgoCD auto-configuration ([a4bbc38](https://github.com/santi1s/yak-gui/commit/a4bbc3801f922aaa7539d2ff5f7f6b26afe5a27c))
* add GitHub Actions workflows and update README for standalone yak-gui ([0f0f1a2](https://github.com/santi1s/yak-gui/commit/0f0f1a2a245f6f454662c172075278f34be41dfd))
* add initial CHANGELOG.md for release-please automation ([3deefd2](https://github.com/santi1s/yak-gui/commit/3deefd2b277e45897ea4b9ace791800fc332c349))
* add SSL certificate management with comprehensive renewal workflow ([6e5bc94](https://github.com/santi1s/yak-gui/commit/6e5bc94c1eb21d506e09c5fa923f83019371c788))
* clean up frontend migration and fix all tests ([8b4e29c](https://github.com/santi1s/yak-gui/commit/8b4e29c7fcb2e2bd44a5e0420ad0bc5aea981ff8))
* improve environment variable handling and fix ArgoCD profile sync ([7f88834](https://github.com/santi1s/yak-gui/commit/7f888348d7fca987f88d979f3341080c15b30b4a))
* release 1.6.0 with comprehensive yak CLI GUI and release-please automation ([ac69bd3](https://github.com/santi1s/yak-gui/commit/ac69bd3acf0ab9220593fe46ed28efbdcb8825a7))
* update window title to match new comprehensive GUI branding ([081f194](https://github.com/santi1s/yak-gui/commit/081f1947a5f9a5af1b495032583bdac6c7992d41))


### Bug Fixes

* add ES module support to frontend package.json ([68fb325](https://github.com/santi1s/yak-gui/commit/68fb3254b022af45b06171f3b6dce42794b46c1a))
* add explicit tag_name to release creation ([94ba7d1](https://github.com/santi1s/yak-gui/commit/94ba7d14f96bf11b1a9fd9760eb09c3a6e784681))
* add missing release-please manifest and update to googleapis action ([db79d79](https://github.com/santi1s/yak-gui/commit/db79d79b6d7a0df9c97be8a28416d96eaace63a4))
* add missing Rollup native module for Linux frontend builds ([44fc4ce](https://github.com/santi1s/yak-gui/commit/44fc4ce0306a28493ac49bc4c91a30bf501088a5))
* configure release-please to only create PRs without approval ([ee638c3](https://github.com/santi1s/yak-gui/commit/ee638c3c9a339b414ec3f66bce29d709e50b7a53))
* implement PATH resolution for macOS GUI and improve rollout image display ([324bd5e](https://github.com/santi1s/yak-gui/commit/324bd5e808c645fd9c00cef2ebca24c0fa2264b4))
* remove complex release-please config and use simple workflow ([adbcb38](https://github.com/santi1s/yak-gui/commit/adbcb38a72c50541ca781f60d37eaca04b2e5ed6))
* remove redundant trigger-build job from release-please ([f476cba](https://github.com/santi1s/yak-gui/commit/f476cba07a6751b9a735c3549af60632259f50d3))
* remove unsupported regex extraFile type from release-please config ([7dee1a1](https://github.com/santi1s/yak-gui/commit/7dee1a114018223a1889c0493987a0003a833477))
* remove unused formatDate function in Secrets.tsx ([b532b1f](https://github.com/santi1s/yak-gui/commit/b532b1f934daf241e0ec6f60e903450055451249))
* remove unused outputStr variable and improve test workflow ([0297910](https://github.com/santi1s/yak-gui/commit/0297910b023ebcdfb1b9db4e4ea7ee3c4f458d37))
* remove unused truncatedOutput variables causing build failures ([073451e](https://github.com/santi1s/yak-gui/commit/073451e2eb02623950dbb68c220e8204064da112))
* rename postcss.config.js to .cjs for ES module compatibility ([aaeacb1](https://github.com/santi1s/yak-gui/commit/aaeacb14f5b032ab3fe8cecd9aa646f29b6afb1e))
* reset versions to 1.5.0 and remove invalid workflow parameter ([097551e](https://github.com/santi1s/yak-gui/commit/097551ef09104cb4202dbe961af7d7f6b537b8ef))
* resolve TypeScript errors in frontend components ([3797cf5](https://github.com/santi1s/yak-gui/commit/3797cf5f3fccbb5054818ab7f77a6e50e558e519))
* restore contents write permission and remove manifest for fresh start ([e74ad83](https://github.com/santi1s/yak-gui/commit/e74ad839b9d12fd5fc542f2fa6dae63c9523dfbf))
* restore manifest and add minimal release-please config ([6d64a52](https://github.com/santi1s/yak-gui/commit/6d64a521ad529fa5a7ef295290a49283f22cfd9c))
* set release-please manifest to current version 1.5.0 ([dc7e100](https://github.com/santi1s/yak-gui/commit/dc7e10057e55854a46a08cb93db8324fd0a0414c))
* simplify release-please config and add bootstrap-sha ([b011337](https://github.com/santi1s/yak-gui/commit/b01133770e0811fcb8b011150a15d142716f61f5))
* update rollout status command to use 'yak rollouts get' instead of 'status' ([9c35a43](https://github.com/santi1s/yak-gui/commit/9c35a43bb76104f05aa2ef8cd964c88652fb0253))
* use 'yak' from PATH instead of relative path '../yak' ([df8ecbf](https://github.com/santi1s/yak-gui/commit/df8ecbf6daaee49ce057ccad9872538d4e17ffbe))

## [1.7.0](https://github.com/santi1s/yak-gui/compare/v1.6.0...v1.7.0) (2025-07-15)


### Features

* improve environment variable handling and fix ArgoCD profile sync ([7f88834](https://github.com/santi1s/yak-gui/commit/7f888348d7fca987f88d979f3341080c15b30b4a))


### Bug Fixes

* add explicit tag_name to release creation ([94ba7d1](https://github.com/santi1s/yak-gui/commit/94ba7d14f96bf11b1a9fd9760eb09c3a6e784681))
* remove redundant trigger-build job from release-please ([f476cba](https://github.com/santi1s/yak-gui/commit/f476cba07a6751b9a735c3549af60632259f50d3))

## [1.6.0](https://github.com/santi1s/yak-gui/compare/v1.5.0...v1.6.0) (2025-07-15)


### Features

* add comprehensive backend testing infrastructure ([1804e0e](https://github.com/santi1s/yak-gui/commit/1804e0e9492412d837e169a50a53d478b627f19a))
* add environment profiles, version display, and ArgoCD auto-configuration ([a4bbc38](https://github.com/santi1s/yak-gui/commit/a4bbc3801f922aaa7539d2ff5f7f6b26afe5a27c))
* add initial CHANGELOG.md for release-please automation ([3deefd2](https://github.com/santi1s/yak-gui/commit/3deefd2b277e45897ea4b9ace791800fc332c349))
* add SSL certificate management with comprehensive renewal workflow ([6e5bc94](https://github.com/santi1s/yak-gui/commit/6e5bc94c1eb21d506e09c5fa923f83019371c788))
* clean up frontend migration and fix all tests ([8b4e29c](https://github.com/santi1s/yak-gui/commit/8b4e29c7fcb2e2bd44a5e0420ad0bc5aea981ff8))
* release 1.6.0 with comprehensive yak CLI GUI and release-please automation ([ac69bd3](https://github.com/santi1s/yak-gui/commit/ac69bd3acf0ab9220593fe46ed28efbdcb8825a7))
* update window title to match new comprehensive GUI branding ([081f194](https://github.com/santi1s/yak-gui/commit/081f1947a5f9a5af1b495032583bdac6c7992d41))


### Bug Fixes

* add missing release-please manifest and update to googleapis action ([db79d79](https://github.com/santi1s/yak-gui/commit/db79d79b6d7a0df9c97be8a28416d96eaace63a4))
* configure release-please to only create PRs without approval ([ee638c3](https://github.com/santi1s/yak-gui/commit/ee638c3c9a339b414ec3f66bce29d709e50b7a53))
* remove complex release-please config and use simple workflow ([adbcb38](https://github.com/santi1s/yak-gui/commit/adbcb38a72c50541ca781f60d37eaca04b2e5ed6))
* remove unsupported regex extraFile type from release-please config ([7dee1a1](https://github.com/santi1s/yak-gui/commit/7dee1a114018223a1889c0493987a0003a833477))
* reset versions to 1.5.0 and remove invalid workflow parameter ([097551e](https://github.com/santi1s/yak-gui/commit/097551ef09104cb4202dbe961af7d7f6b537b8ef))
* restore contents write permission and remove manifest for fresh start ([e74ad83](https://github.com/santi1s/yak-gui/commit/e74ad839b9d12fd5fc542f2fa6dae63c9523dfbf))
* restore manifest and add minimal release-please config ([6d64a52](https://github.com/santi1s/yak-gui/commit/6d64a521ad529fa5a7ef295290a49283f22cfd9c))
* set release-please manifest to current version 1.5.0 ([dc7e100](https://github.com/santi1s/yak-gui/commit/dc7e10057e55854a46a08cb93db8324fd0a0414c))
* simplify release-please config and add bootstrap-sha ([b011337](https://github.com/santi1s/yak-gui/commit/b01133770e0811fcb8b011150a15d142716f61f5))

## [1.5.0] - 2024-12-19

### Features
- Add SSL certificate management with comprehensive renewal workflow
- Add environment profiles, version display, and ArgoCD auto-configuration
- Update rollout status command to use 'yak rollouts get' instead of 'status'

### Bug Fixes
- Various bug fixes and improvements

### Build System
- Update version to 1.5.0
