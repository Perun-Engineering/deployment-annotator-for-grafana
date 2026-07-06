# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- semantic-release-generated changelog -->

## [1.5.10](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.9...v1.5.10) (2026-07-06)


### 📦 Dependencies

* **deps:** bump golang.org/x/text from 0.37.0 to 0.38.0 ([8f37cf4](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/8f37cf4600dbfc894ede4dba3959efcba2a5b08a))
* **deps:** bump k8s.io/api from 0.36.1 to 0.36.2 ([8dae92d](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/8dae92d4ed3af4f2bed7fdeec18c0879468bdc57))
* **deps:** bump k8s.io/apimachinery from 0.36.1 to 0.36.2 ([cf1fa61](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/cf1fa6140091ae9d09ff0028c44e5ab501de78aa))

## [1.5.9](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.8...v1.5.9) (2026-06-14)


### 📦 Dependencies

* **deps:** bump k8s.io/api from 0.36.0 to 0.36.1 ([#113](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/113)) ([57616a9](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/57616a948130fb2e1e5e980741f803a6bbd6638f))

## [1.5.8](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.7...v1.5.8) (2026-05-31)


### 📚 Documentation

* add agent skills config and first ADR ([#110](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/110)) ([762fc0e](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/762fc0e2b64bba472a1942c9d9619535d6313556))

## [1.5.7](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.6...v1.5.7) (2026-05-31)


### 📦 Dependencies

* **deps:** bump k8s.io/apimachinery from 0.36.0 to 0.36.1 ([658fb4d](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/658fb4d82130244d143d1196bdd00654b0001de6))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.23.3 to 0.24.1 ([#108](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/108)) ([69178ff](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/69178ff8c8ad27a39c00e974a7b10d21d5ec63a5))

## [1.5.6](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.5...v1.5.6) (2026-05-31)


### 📦 Dependencies

* **deps:** bump go.uber.org/zap from 1.27.1 to 1.28.0 ([#105](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/105)) ([f33c9d7](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/f33c9d7d1be660d555c374e5aa90674262433fec))
* **deps:** bump golang.org/x/text from 0.36.0 to 0.37.0 ([#107](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/107)) ([8d79c78](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/8d79c780e53013cf50b2bde8ae1f6a107df51bec))

## [1.5.5](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/compare/v1.5.4...v1.5.5) (2026-05-01)


### 🐛 Bug Fixes

* **ci:** configure semantic-release authentication for workflow_run trigger ([bd289a7](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/bd289a7c8afa1cf56acc60a4a0fa586c68c964de))
* **ci:** hardcode checkout ref to prevent workflow_run code injection ([5570c4b](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/5570c4bc38dd3c0a1fe3f6b58eef6b522b9c1f8e))
* **ci:** remove repositoryUrl to use checkout credentials for git push ([407edc1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/407edc1c97d513f87035c8158c70db527e1221aa))
* **ci:** update Go version to 1.25 ([a18eef5](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/a18eef5be21468a1b18088c3124cd600a17c52c0))


### ♻️ Code Refactoring

* absorb extractContainerImage into WorkloadAdapter interface ([947a8e9](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/947a8e92173ea76d03fa712762727b8672633427))
* dissolve internal/util package into internal/controller ([4fe62f2](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/4fe62f2f2fc44c7e4f263911269f7199b27df821))
* extract AnnotationClient interface seam from concrete grafana.Client ([bfb2e4c](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/bfb2e4cae5c0c2a0c7aaacc78e37842ea0d1cc2d))
* extract AnnotationLifecycle and complete adapter seam ([#102](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/102)) ([7e0a695](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/7e0a695fe63e956ceb5ddb915ed89a84fd6e3e0d))
* inject time seam into grafana.Client for testability ([1fec16c](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/1fec16c47ccafed701e36d8d9fe455eecf71d623))
* unify three reconcilers into single WorkloadReconciler with adapter pattern ([e9a3ad1](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/e9a3ad1deed5cd5f7ea82729ba1cd572dbbefa66))


### 📦 Dependencies

* bump k8s.io/{api,apimachinery,client-go} to v0.35.4 ([#101](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/101)) ([dfd29fb](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/dfd29fb9c7f6eaf6b7dd4b3ecdb12eeb8cd9d182)), closes [#99](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/99) [#100](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/100) [#98](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/98) [#99](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/99) [#100](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/issues/100)
* **deps:** bump go.uber.org/zap from 1.27.0 to 1.27.1 ([4a4981b](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/4a4981b800423c8497f499f4f2a4297acd65b06b))
* **deps:** bump golang from 1.25-alpine to 1.26-alpine ([e96a1d5](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/e96a1d5e5774023e9c3cb3a4b31b9f554d356561))
* **deps:** bump golang.org/x/text from 0.30.0 to 0.32.0 ([c27f31d](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/c27f31df8cf8b54c5eb5b43446f6d3135fd76147))
* **deps:** bump golang.org/x/text from 0.32.0 to 0.34.0 ([2b8665e](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/2b8665e6b7bd0983bd78fcc0f4f4f5c1ae68a6e5))
* **deps:** bump golang.org/x/text from 0.34.0 to 0.36.0 ([554edb4](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/554edb46ba330e961d1e37c5948703fe3d0cab91))
* **deps:** bump k8s.io/api from 0.34.1 to 0.34.3 ([2ebbe98](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/2ebbe985f28d13206fabd7e34895b3325735ec3e))
* **deps:** bump k8s.io/api from 0.35.0 to 0.35.2 ([edbed97](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/edbed979d0265c02c46423c80a3a6b62403dedbb))
* **deps:** bump k8s.io/apimachinery from 0.34.1 to 0.34.3 ([ba8675d](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/ba8675dc0fa13f95bbc3c74a948e72791ff775aa))
* **deps:** bump k8s.io/apimachinery from 0.35.0 to 0.35.2 ([c7a688c](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/c7a688c4af237a5dd22e3cd5e1e546bd3d16c61a))
* **deps:** bump k8s.io/client-go from 0.34.1 to 0.34.3 ([a45148f](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/a45148f4e17a1e3a6739207b9a6fb4a6e82ddd27))
* **deps:** bump k8s.io/client-go from 0.34.3 to 0.35.0 ([026482b](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/026482b18a2bd8614253a8e845c641dc40c04411))
* **deps:** bump k8s.io/client-go from 0.35.0 to 0.35.2 ([4f6831f](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/4f6831f79a0e33f47a0d84012d01bb839a6ea846))
* **deps:** bump k8s.io/client-go from 0.35.2 to 0.35.3 ([aa618eb](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/aa618eb873c31751105159daa08a2bf7df1d47fa))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.22.3 to 0.22.4 ([dd06c80](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/dd06c80b47fa64b104de93d0f9d0bfa21db5506a))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.22.4 to 0.23.1 ([f50fdbf](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/f50fdbfc3cca3b5ecb135d034e36a5157c2fd3d5))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.23.1 to 0.23.3 ([325f805](https://github.com/Perun-Engineering/deployment-annotator-for-grafana/commit/325f805f60a7678fcce1c3453b1a910e0f4d1216))

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
