# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- semantic-release-generated changelog -->

## [1.5.4](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.3...v1.5.4) (2025-10-17)


### 📦 Dependencies

* **deps:** bump golang.org/x/text from 0.29.0 to 0.30.0 ([2247be1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/2247be1b0b3d522038cf0bd94e078fa3186c3960))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.22.1 to 0.22.3 ([5fb8af3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/5fb8af320dc1bf5de3428ef2955fc7ba31bf1bd6))

## [1.5.3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.2...v1.5.3) (2025-09-11)


### 📦 Dependencies

* **deps:** bump golang.org/x/text from 0.28.0 to 0.29.0 ([bba23b6](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/bba23b61630afa24ff924f3bb2f57e7267a5de51))
* **deps:** bump k8s.io/apimachinery from 0.33.4 to 0.34.0 ([c978ef4](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/c978ef4301d30effe0cc031d6d27c6b6fbcb8851))
* **deps:** bump k8s.io/client-go from 0.33.4 to 0.34.0 ([7756e8b](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/7756e8b5497a6b13e2a6f418db1692deb44b2198))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.21.0 to 0.22.0 ([7112669](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/711266909ba0083257cf94c3ab45784d1c498de8))

## [1.5.2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.1...v1.5.2) (2025-08-19)


### 📦 Dependencies

* **deps:** bump golang from 1.24-alpine to 1.25-alpine ([40cc0c2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/40cc0c2a5d379faf7229b9b9fd4c44ca20f8cbd2))

## [1.5.1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.0...v1.5.1) (2025-08-19)


### 📦 Dependencies

* **deps:** bump k8s.io/client-go from 0.33.3 to 0.33.4 ([9769812](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/97698124d544be4fa7d9dd336e2f93106a660633))

## [1.5.0](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.4.1...v1.5.0) (2025-08-08)


### 🎯 Features

* add selective workload watching ([ad6ee81](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/ad6ee8145d6ae494c917123bb3be6216899278df))

## [1.4.1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.4.0...v1.4.1) (2025-08-08)


### 🐛 Bug Fixes

* refactor codebase to be more modular ([115778d](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/115778d3dce6133ee5eed0b818c353fa8654524c))


### ♻️ Code Refactoring

* add internal grafana client and util ([f12c320](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/f12c320e8c1da13b8db5afec38e4c02a9de1afdc))

## [1.4.0](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.3.2...v1.4.0) (2025-08-08)


### 🎯 Features

* Added support for StatefulSets and DaemonSets ([997a890](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/997a890d6e26b091b77a7a18fed67e3d01e1464f))

## [1.3.2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.3.1...v1.3.2) (2025-08-05)


### 📦 Dependencies

* **deps:** bump k8s.io/apimachinery from 0.33.2 to 0.33.3 ([3cf6476](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/3cf647687d57cc526920eb27cd2451fbb1a50d34))
* **deps:** bump k8s.io/client-go from 0.33.2 to 0.33.3 ([9e6dc43](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/9e6dc43347f981f975930e701db944ad467d1368))

## [1.3.1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.3.0...v1.3.1) (2025-08-05)


### 🐛 Bug Fixes

* add namespace label check for deployment deletions ([24685ef](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/24685ef1a1e82f7b1fecae16803c246248eda93c))

## [1.3.0](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.2.2...v1.3.0) (2025-08-04)


### 🎯 Features

* **grafana:** add connection test and refactor client ([d806089](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/d8060896667823964c01b585010ebea91d17bf7c))

## [1.2.2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.2.1...v1.2.2) (2025-06-27)


### 🐛 Bug Fixes

* move deployment predicate to For() to allow namespace events ([c680f36](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/c680f36b5ab8f2cef5669b9f954925532dc20b85))

## [1.2.1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.2.0...v1.2.1) (2025-06-27)


### 📚 Documentation

* update README for namespace watching and new features ([95e2ecf](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/95e2ecf755bee28af3d31b7e4582a57430513c8f))

## [1.2.0](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.1.0...v1.2.0) (2025-06-27)


### 🎯 Features

* add namespace watching with annotation cleanup ([2356440](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/235644066437408c42c98e0024271b5f46cd9eed))

## [1.1.0](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.7...v1.1.0) (2025-06-27)


### 🎯 Features

* add configurable logging and fix annotation handling ([6c0673f](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/6c0673f913cf286e1baf4694d014123bc62b1f8f))

## [1.0.7](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.6...v1.0.7) (2025-06-25)


### 🐛 Bug Fixes

* use ReplicaSet pod template hash to prevent feedback loops ([02f08d2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/02f08d24d56ab2520826401f0327340f8fb3ee4b))

## [1.0.6](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.5...v1.0.6) (2025-06-25)


### 🐛 Bug Fixes

* implement proper deployment tracking initialization ([73af7d3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/73af7d3e4a6956df82180ab66b21f886f67ca062))

## [1.0.5](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.4...v1.0.5) (2025-06-25)


### 🐛 Bug Fixes

* **k8s:** add missing ReplicaSet permissions to RBAC ([7c2f7f3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/7c2f7f353dc38f497d7de091f341dca274ef7fa9))

## [1.0.4](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.3...v1.0.4) (2025-06-25)


### 📚 Documentation

* remove references to deleted examples folder ([f81196b](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/f81196bd2c9d69ff51010b6e6d9ab04f0dfdc9fe))

## [1.0.3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.2...v1.0.3) (2025-06-24)


### 📚 Documentation

* improve README documentation ([d09cf28](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/d09cf286090c04a2c8f785b95ae87ed09b2d6718))

## [1.0.2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.1...v1.0.2) (2025-06-24)


### 🐛 Bug Fixes

* **ci:** more fixes to the helm chart publishing ([32df87f](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/32df87fe05de4031335d4598cc7f02a51351ba9c))

## [1.0.1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.0.0...v1.0.1) (2025-06-24)


### 🐛 Bug Fixes

* **ci:** fix path to the helm chart ([6ba0d8a](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/6ba0d8ac9dac9b9e669d4231be08800b6e0b1cca))

## 1.0.0 (2025-06-24)


### 🎯 Features

* **ci:** disable PR labels validation ([05846b3](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/05846b389ac8ec23ac9d2748cc2312b9c8ec221b))


### 🐛 Bug Fixes

* **ci:** configure semantic-release for branch protection bypass ([f698d38](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/f698d3874057222fcf77b2d3d00c7cb1b8a92848))
* **ci:** Docker tag generation with invalid leading dash ([bf7f719](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/bf7f719e03f98b093534615368d9154b8b3757b0))
* **ci:** multi-arch build execution for workflow_run events ([d7c20cc](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/d7c20cc288a6b04055a203368b7721f9f8efb2bc))
* **ci:** multi-arch workflow logic for pull requests ([3a1127e](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/3a1127e8425639fa88718db0c2ab9c59bcc03c81))
* **ci:** resolve critical workflow issues preventing pipeline execution ([98ddf65](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/98ddf658405d0becf42ef59bc793441495d8c1dd))
