# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
