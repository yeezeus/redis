# Change Log

## [Unreleased](https://github.com/kubedb/redis/tree/HEAD)

[Full Changelog](https://github.com/kubedb/redis/compare/v0.6.0-rc.0...HEAD)

**Merged pull requests:**

- Support PodAffinity Templating [\#155](https://github.com/kubedb/redis/pull/155) ([faem](https://github.com/faem))
- Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 [\#154](https://github.com/kubedb/redis/pull/154) ([tamalsaha](https://github.com/tamalsaha))
- Version update to resolve security issue in github.com/apache/thrift [\#153](https://github.com/kubedb/redis/pull/153) ([MitaliBo](https://github.com/MitaliBo))
- Use rancher/local-path-provisioner@v0.0.12 [\#152](https://github.com/kubedb/redis/pull/152) ([tamalsaha](https://github.com/tamalsaha))
- Introduce spec.halted and removed dormant crd [\#151](https://github.com/kubedb/redis/pull/151) ([the-redback](https://github.com/the-redback))
- Add `Pause` Feature [\#150](https://github.com/kubedb/redis/pull/150) ([faem](https://github.com/faem))
- Refactor CI pipeline to build once [\#149](https://github.com/kubedb/redis/pull/149) ([tamalsaha](https://github.com/tamalsaha))
- Update kubernetes client-go to 1.16.3 [\#148](https://github.com/kubedb/redis/pull/148) ([tamalsaha](https://github.com/tamalsaha))
- Update catalog values for make install command [\#147](https://github.com/kubedb/redis/pull/147) ([tamalsaha](https://github.com/tamalsaha))
- Use charts to install operator [\#146](https://github.com/kubedb/redis/pull/146) ([tamalsaha](https://github.com/tamalsaha))
- Matrix test for github actions [\#145](https://github.com/kubedb/redis/pull/145) ([the-redback](https://github.com/the-redback))
- Add license header to files [\#144](https://github.com/kubedb/redis/pull/144) ([tamalsaha](https://github.com/tamalsaha))
- Use log.Fatal instead of Must\(\) [\#143](https://github.com/kubedb/redis/pull/143) ([shudipta](https://github.com/shudipta))
- Run e2e tests in parallel [\#142](https://github.com/kubedb/redis/pull/142) ([tamalsaha](https://github.com/tamalsaha))
- Enable make ci [\#141](https://github.com/kubedb/redis/pull/141) ([tamalsaha](https://github.com/tamalsaha))
- Remove EnableStatusSubresource [\#140](https://github.com/kubedb/redis/pull/140) ([tamalsaha](https://github.com/tamalsaha))
- Fix tests for github actions [\#139](https://github.com/kubedb/redis/pull/139) ([shudipta](https://github.com/shudipta))
- Validate DBVersionSpecs and fixed broken build [\#138](https://github.com/kubedb/redis/pull/138) ([the-redback](https://github.com/the-redback))
- Run e2e tests using GitHub actions [\#137](https://github.com/kubedb/redis/pull/137) ([tamalsaha](https://github.com/tamalsaha))
- Prepend redis.conf to the arguments for redis-server cmd [\#136](https://github.com/kubedb/redis/pull/136) ([shudipta](https://github.com/shudipta))

## [v0.6.0-rc.0](https://github.com/kubedb/redis/tree/v0.6.0-rc.0) (2019-08-22)
[Full Changelog](https://github.com/kubedb/redis/compare/0.5.0...v0.6.0-rc.0)

**Merged pull requests:**

- Improve test: Use installed redisversions [\#135](https://github.com/kubedb/redis/pull/135) ([the-redback](https://github.com/the-redback))
- Update dependencies [\#134](https://github.com/kubedb/redis/pull/134) ([tamalsaha](https://github.com/tamalsaha))
- Don't set annotation to AppBinding [\#133](https://github.com/kubedb/redis/pull/133) ([hossainemruz](https://github.com/hossainemruz))
- Set database version in AppBinding [\#132](https://github.com/kubedb/redis/pull/132) ([hossainemruz](https://github.com/hossainemruz))
- Change package path to kubedb.dev/redis [\#131](https://github.com/kubedb/redis/pull/131) ([tamalsaha](https://github.com/tamalsaha))
- Add license header to Makefiles [\#130](https://github.com/kubedb/redis/pull/130) ([tamalsaha](https://github.com/tamalsaha))
- Add install, uninstall and purge command in Makefile [\#129](https://github.com/kubedb/redis/pull/129) ([hossainemruz](https://github.com/hossainemruz))
- Add Makefile [\#128](https://github.com/kubedb/redis/pull/128) ([tamalsaha](https://github.com/tamalsaha))
- Pod Disruption Budget for Redis [\#127](https://github.com/kubedb/redis/pull/127) ([iamrz1](https://github.com/iamrz1))
- Handling resource ownership [\#126](https://github.com/kubedb/redis/pull/126) ([iamrz1](https://github.com/iamrz1))
- Update to k8s 1.14.0 client libraries using go.mod [\#125](https://github.com/kubedb/redis/pull/125) ([tamalsaha](https://github.com/tamalsaha))

## [0.5.0](https://github.com/kubedb/redis/tree/0.5.0) (2019-05-06)
[Full Changelog](https://github.com/kubedb/redis/compare/0.4.0...0.5.0)

**Merged pull requests:**

- Revendor dependencies [\#124](https://github.com/kubedb/redis/pull/124) ([tamalsaha](https://github.com/tamalsaha))
- Fix PSP in Role for kubeDB operator upgrade [\#123](https://github.com/kubedb/redis/pull/123) ([iamrz1](https://github.com/iamrz1))
- Modify mutator validator names [\#122](https://github.com/kubedb/redis/pull/122) ([iamrz1](https://github.com/iamrz1))

## [0.4.0](https://github.com/kubedb/redis/tree/0.4.0) (2019-03-18)
[Full Changelog](https://github.com/kubedb/redis/compare/0.3.0...0.4.0)

**Merged pull requests:**

- DB psp in e2e test framework [\#121](https://github.com/kubedb/redis/pull/121) ([iamrz1](https://github.com/iamrz1))
- Add clustering support for Redis v5 [\#120](https://github.com/kubedb/redis/pull/120) ([shudipta](https://github.com/shudipta))
- Don't inherit app.kubernetes.io labels from CRD into offshoots [\#119](https://github.com/kubedb/redis/pull/119) ([tamalsaha](https://github.com/tamalsaha))
- Add role label to stats service [\#118](https://github.com/kubedb/redis/pull/118) ([tamalsaha](https://github.com/tamalsaha))
- Fix bugs for unknown nodes in nodes.conf [\#117](https://github.com/kubedb/redis/pull/117) ([shudipta](https://github.com/shudipta))
- PSP support for Redis. [\#116](https://github.com/kubedb/redis/pull/116) ([iamrz1](https://github.com/iamrz1))
- Update Kubernetes client libraries to 1.13.0 release [\#115](https://github.com/kubedb/redis/pull/115) ([tamalsaha](https://github.com/tamalsaha))

## [0.3.0](https://github.com/kubedb/redis/tree/0.3.0) (2019-02-19)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0...0.3.0)

**Merged pull requests:**

- Revendor dependencies [\#114](https://github.com/kubedb/redis/pull/114) ([tamalsaha](https://github.com/tamalsaha))
- Initial RBAC support: create and use K8s service account for Redis [\#112](https://github.com/kubedb/redis/pull/112) ([the-redback](https://github.com/the-redback))
- Revendor dependencies [\#111](https://github.com/kubedb/redis/pull/111) ([the-redback](https://github.com/the-redback))
- Add certificate health checker [\#110](https://github.com/kubedb/redis/pull/110) ([tamalsaha](https://github.com/tamalsaha))
- Redis 5.0.3, 5.0 Support added [\#109](https://github.com/kubedb/redis/pull/109) ([the-redback](https://github.com/the-redback))
- Update E2E test: Env update is not restricted anymore [\#108](https://github.com/kubedb/redis/pull/108) ([the-redback](https://github.com/the-redback))
- Fix AppBinding [\#107](https://github.com/kubedb/redis/pull/107) ([tamalsaha](https://github.com/tamalsaha))
- Redis cluster support [\#94](https://github.com/kubedb/redis/pull/94) ([shudipta](https://github.com/shudipta))

## [0.2.0](https://github.com/kubedb/redis/tree/0.2.0) (2018-12-17)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0-rc.2...0.2.0)

**Merged pull requests:**

- Reuse event recorder [\#106](https://github.com/kubedb/redis/pull/106) ([tamalsaha](https://github.com/tamalsaha))
- Revendor dependencies [\#105](https://github.com/kubedb/redis/pull/105) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-rc.2](https://github.com/kubedb/redis/tree/0.2.0-rc.2) (2018-12-06)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0-rc.1...0.2.0-rc.2)

**Merged pull requests:**

- Ignore mutation of fields to default values during update [\#103](https://github.com/kubedb/redis/pull/103) ([tamalsaha](https://github.com/tamalsaha))
- Support configuration options for exporter sidecar [\#102](https://github.com/kubedb/redis/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Use flags.DumpAll [\#101](https://github.com/kubedb/redis/pull/101) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-rc.1](https://github.com/kubedb/redis/tree/0.2.0-rc.1) (2018-12-02)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0-rc.0...0.2.0-rc.1)

**Merged pull requests:**

- Apply cleanup [\#100](https://github.com/kubedb/redis/pull/100) ([tamalsaha](https://github.com/tamalsaha))
- Set periodic analytics [\#99](https://github.com/kubedb/redis/pull/99) ([tamalsaha](https://github.com/tamalsaha))
-  Tests for appbinding [\#98](https://github.com/kubedb/redis/pull/98) ([the-redback](https://github.com/the-redback))
- Fix analytics [\#97](https://github.com/kubedb/redis/pull/97) ([the-redback](https://github.com/the-redback))
- Fix operator startup in minikube [\#96](https://github.com/kubedb/redis/pull/96) ([the-redback](https://github.com/the-redback))
- Add CRDS without observation when operator starts [\#95](https://github.com/kubedb/redis/pull/95) ([the-redback](https://github.com/the-redback))

## [0.2.0-rc.0](https://github.com/kubedb/redis/tree/0.2.0-rc.0) (2018-10-15)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0-beta.1...0.2.0-rc.0)

**Merged pull requests:**

- Support providing resources for monitoring container [\#93](https://github.com/kubedb/redis/pull/93) ([tamalsaha](https://github.com/tamalsaha))
- Update kubernetes client libraries to 1.12.0 [\#92](https://github.com/kubedb/redis/pull/92) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook xray [\#91](https://github.com/kubedb/redis/pull/91) ([tamalsaha](https://github.com/tamalsaha))
- Various Fixes [\#90](https://github.com/kubedb/redis/pull/90) ([hossainemruz](https://github.com/hossainemruz))
- Merge ports from service template [\#89](https://github.com/kubedb/redis/pull/89) ([tamalsaha](https://github.com/tamalsaha))
- Remove remaining DoNotPause [\#88](https://github.com/kubedb/redis/pull/88) ([tamalsaha](https://github.com/tamalsaha))
- Replace doNotPause with TerminationPolicy = DoNotTerminate [\#87](https://github.com/kubedb/redis/pull/87) ([tamalsaha](https://github.com/tamalsaha))
- Pass resources to NamespaceValidator [\#85](https://github.com/kubedb/redis/pull/85) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook for Namespace deletion [\#84](https://github.com/kubedb/redis/pull/84) ([tamalsaha](https://github.com/tamalsaha))
- Use FQDN for kube-apiserver in AKS [\#83](https://github.com/kubedb/redis/pull/83) ([tamalsaha](https://github.com/tamalsaha))
- Support Livecycle hook and container probes [\#82](https://github.com/kubedb/redis/pull/82) ([tamalsaha](https://github.com/tamalsaha))
- Check if Kubernetes version is supported before running operator [\#81](https://github.com/kubedb/redis/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- Update package alias [\#80](https://github.com/kubedb/redis/pull/80) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-beta.1](https://github.com/kubedb/redis/tree/0.2.0-beta.1) (2018-09-30)
[Full Changelog](https://github.com/kubedb/redis/compare/0.2.0-beta.0...0.2.0-beta.1)

**Merged pull requests:**

- Revendor api [\#79](https://github.com/kubedb/redis/pull/79) ([tamalsaha](https://github.com/tamalsaha))
- Fix tests [\#78](https://github.com/kubedb/redis/pull/78) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api for catalog apigroup [\#77](https://github.com/kubedb/redis/pull/77) ([tamalsaha](https://github.com/tamalsaha))
- Use --pull flag with docker build \(\#20\) [\#76](https://github.com/kubedb/redis/pull/76) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-beta.0](https://github.com/kubedb/redis/tree/0.2.0-beta.0) (2018-09-20)
[Full Changelog](https://github.com/kubedb/redis/compare/0.1.0...0.2.0-beta.0)

**Merged pull requests:**

- Support Termination Policy & Stop working for deprecated \*Versions [\#73](https://github.com/kubedb/redis/pull/73) ([the-redback](https://github.com/the-redback))
- Revendor k8s.io/apiserver [\#72](https://github.com/kubedb/redis/pull/72) ([tamalsaha](https://github.com/tamalsaha))
- Revendor kubernetes-1.11.3 [\#71](https://github.com/kubedb/redis/pull/71) ([tamalsaha](https://github.com/tamalsaha))
- Support UpdateStrategy [\#70](https://github.com/kubedb/redis/pull/70) ([tamalsaha](https://github.com/tamalsaha))
- Add TerminationPolicy for databases [\#69](https://github.com/kubedb/redis/pull/69) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#68](https://github.com/kubedb/redis/pull/68) ([tamalsaha](https://github.com/tamalsaha))
- Use IntHash as status.observedGeneration [\#67](https://github.com/kubedb/redis/pull/67) ([tamalsaha](https://github.com/tamalsaha))
- fix github status [\#66](https://github.com/kubedb/redis/pull/66) ([tahsinrahman](https://github.com/tahsinrahman))
- update pipeline [\#65](https://github.com/kubedb/redis/pull/65) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix E2E test for minikube [\#64](https://github.com/kubedb/redis/pull/64) ([the-redback](https://github.com/the-redback))
- update pipeline [\#63](https://github.com/kubedb/redis/pull/63) ([tahsinrahman](https://github.com/tahsinrahman))
- Use officially suggested exporter image [\#62](https://github.com/kubedb/redis/pull/62) ([the-redback](https://github.com/the-redback))
- Migrate Redis [\#61](https://github.com/kubedb/redis/pull/61) ([tamalsaha](https://github.com/tamalsaha))
- Update status.ObservedGeneration for failure phase [\#60](https://github.com/kubedb/redis/pull/60) ([the-redback](https://github.com/the-redback))
- Keep track of ObservedGenerationHash [\#59](https://github.com/kubedb/redis/pull/59) ([tamalsaha](https://github.com/tamalsaha))
- Use NewObservableHandler [\#58](https://github.com/kubedb/redis/pull/58) ([tamalsaha](https://github.com/tamalsaha))
- Fix uninstall for concourse [\#57](https://github.com/kubedb/redis/pull/57) ([tahsinrahman](https://github.com/tahsinrahman))
- Revise immutable spec fields [\#56](https://github.com/kubedb/redis/pull/56) ([tamalsaha](https://github.com/tamalsaha))
- Support passing args via PodTemplate [\#55](https://github.com/kubedb/redis/pull/55) ([tamalsaha](https://github.com/tamalsaha))
- Introduce storageType : ephemeral [\#54](https://github.com/kubedb/redis/pull/54) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#53](https://github.com/kubedb/redis/pull/53) ([tamalsaha](https://github.com/tamalsaha))
- Add support for running tests on cncf cluster [\#52](https://github.com/kubedb/redis/pull/52) ([tahsinrahman](https://github.com/tahsinrahman))
- Keep track of observedGeneration in status [\#51](https://github.com/kubedb/redis/pull/51) ([tamalsaha](https://github.com/tamalsaha))
- Separate StatsService for monitoring [\#50](https://github.com/kubedb/redis/pull/50) ([shudipta](https://github.com/shudipta))
- Use RedisVersion for Redis images [\#49](https://github.com/kubedb/redis/pull/49) ([shudipta](https://github.com/shudipta))
- Use updated crd spec [\#48](https://github.com/kubedb/redis/pull/48) ([tamalsaha](https://github.com/tamalsaha))
- Rename OffshootLabels to OffshootSelectors [\#47](https://github.com/kubedb/redis/pull/47) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#46](https://github.com/kubedb/redis/pull/46) ([tamalsaha](https://github.com/tamalsaha))
- Use kmodules monitoring and objectstore api [\#45](https://github.com/kubedb/redis/pull/45) ([tamalsaha](https://github.com/tamalsaha))
- Refactor concourse scripts [\#44](https://github.com/kubedb/redis/pull/44) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix command `./hack/make.py test e2e` [\#43](https://github.com/kubedb/redis/pull/43) ([the-redback](https://github.com/the-redback))
- Support custom configuration [\#42](https://github.com/kubedb/redis/pull/42) ([hossainemruz](https://github.com/hossainemruz))
- Don't add admission group as a prioritized version [\#41](https://github.com/kubedb/redis/pull/41) ([tamalsaha](https://github.com/tamalsaha))
- Set generated binary name to rd-operator [\#40](https://github.com/kubedb/redis/pull/40) ([tamalsaha](https://github.com/tamalsaha))
- Format shell script [\#39](https://github.com/kubedb/redis/pull/39) ([tamalsaha](https://github.com/tamalsaha))
- Enable status subresource for crds [\#38](https://github.com/kubedb/redis/pull/38) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v8.0.0 [\#37](https://github.com/kubedb/redis/pull/37) ([tamalsaha](https://github.com/tamalsaha))
- Support ENV variables in CRDs [\#36](https://github.com/kubedb/redis/pull/36) ([hossainemruz](https://github.com/hossainemruz))

## [0.1.0](https://github.com/kubedb/redis/tree/0.1.0) (2018-06-12)
[Full Changelog](https://github.com/kubedb/redis/compare/0.1.0-rc.0...0.1.0)

**Merged pull requests:**

- Fix missing error return [\#35](https://github.com/kubedb/redis/pull/35) ([the-redback](https://github.com/the-redback))
- Revendor dependencies [\#32](https://github.com/kubedb/redis/pull/32) ([tamalsaha](https://github.com/tamalsaha))
- Rename docker build script [\#31](https://github.com/kubedb/redis/pull/31) ([tamalsaha](https://github.com/tamalsaha))
- Add changelog [\#30](https://github.com/kubedb/redis/pull/30) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.0](https://github.com/kubedb/redis/tree/0.1.0-rc.0) (2018-05-28)
[Full Changelog](https://github.com/kubedb/redis/compare/0.1.0-beta.2...0.1.0-rc.0)

**Merged pull requests:**

- Fixed kubeconfig plugin for Cloud Providers && Storage is required for Redis [\#29](https://github.com/kubedb/redis/pull/29) ([the-redback](https://github.com/the-redback))
- Concourse [\#28](https://github.com/kubedb/redis/pull/28) ([tahsinrahman](https://github.com/tahsinrahman))
- Refactored E2E testing to support E2E testing with admission webhook in cloud [\#27](https://github.com/kubedb/redis/pull/27) ([the-redback](https://github.com/the-redback))
- Skip delete requests for empty resources [\#26](https://github.com/kubedb/redis/pull/26) ([the-redback](https://github.com/the-redback))
- Don't panic if admission options is nil [\#25](https://github.com/kubedb/redis/pull/25) ([tamalsaha](https://github.com/tamalsaha))
- Disable admission controllers for webhook server [\#24](https://github.com/kubedb/redis/pull/24) ([tamalsaha](https://github.com/tamalsaha))
- Update Prometheus operator dependency [\#23](https://github.com/kubedb/redis/pull/23) ([tamalsaha](https://github.com/tamalsaha))
- Separate ApiGroup for Mutating and Validating webhook  [\#22](https://github.com/kubedb/redis/pull/22) ([the-redback](https://github.com/the-redback))
- Update client-go to 7.0.0 [\#21](https://github.com/kubedb/redis/pull/21) ([tamalsaha](https://github.com/tamalsaha))
-  Bundle webhook server and used shared Index informer [\#20](https://github.com/kubedb/redis/pull/20) ([the-redback](https://github.com/the-redback))
-  Moved admission webhook packages into redis repository [\#19](https://github.com/kubedb/redis/pull/19) ([the-redback](https://github.com/the-redback))
- Add travis yaml [\#18](https://github.com/kubedb/redis/pull/18) ([tahsinrahman](https://github.com/tahsinrahman))

## [0.1.0-beta.2](https://github.com/kubedb/redis/tree/0.1.0-beta.2) (2018-02-27)
[Full Changelog](https://github.com/kubedb/redis/compare/0.1.0-beta.1...0.1.0-beta.2)

**Merged pull requests:**

- update validation [\#15](https://github.com/kubedb/redis/pull/15) ([aerokite](https://github.com/aerokite))
- Fix dormantDB matching: pass same type to Equal method [\#14](https://github.com/kubedb/redis/pull/14) ([the-redback](https://github.com/the-redback))
- Use official code generator scripts [\#13](https://github.com/kubedb/redis/pull/13) ([tamalsaha](https://github.com/tamalsaha))
- Fixed dormantdb matching & Raised throttling time & Fixed Redis versionn checking [\#12](https://github.com/kubedb/redis/pull/12) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.1](https://github.com/kubedb/redis/tree/0.1.0-beta.1) (2018-01-29)
[Full Changelog](https://github.com/kubedb/redis/compare/0.1.0-beta.0...0.1.0-beta.1)

**Merged pull requests:**

- Update dependencies to client-go v6.0.0 [\#11](https://github.com/kubedb/redis/pull/11) ([tamalsaha](https://github.com/tamalsaha))
-  Fixed logger and analytics [\#10](https://github.com/kubedb/redis/pull/10) ([the-redback](https://github.com/the-redback))
- Reviewed docker images and fixed Monitoring Management [\#9](https://github.com/kubedb/redis/pull/9) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.0](https://github.com/kubedb/redis/tree/0.1.0-beta.0) (2018-01-07)
**Merged pull requests:**

- Fix Analytics and rbac [\#8](https://github.com/kubedb/redis/pull/8) ([the-redback](https://github.com/the-redback))
- Add workqueue & docker-registry flag [\#6](https://github.com/kubedb/redis/pull/6) ([the-redback](https://github.com/the-redback))
- Set client id for analytics [\#5](https://github.com/kubedb/redis/pull/5) ([tamalsaha](https://github.com/tamalsaha))
- Fix CRD registration [\#4](https://github.com/kubedb/redis/pull/4) ([the-redback](https://github.com/the-redback))
- Update pkg paths to kubedb org [\#3](https://github.com/kubedb/redis/pull/3) ([tamalsaha](https://github.com/tamalsaha))
-  Assign default Prometheus Monitoring Port [\#2](https://github.com/kubedb/redis/pull/2) ([the-redback](https://github.com/the-redback))
- Initial implementation [\#1](https://github.com/kubedb/redis/pull/1) ([the-redback](https://github.com/the-redback))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*