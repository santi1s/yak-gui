# Changelog

## [2.23.0](https://github.com/doctolib/yak/compare/v2.22.2...v2.23.0) (2025-07-11)


### Features

* **PSRE-4702:** Add  commands to manage Argo-rollouts ([#1162](https://github.com/doctolib/yak/issues/1162)) ([d5a90bc](https://github.com/doctolib/yak/commit/d5a90bcd614f7b5c78063dc869617abc594aefab))

## [2.22.2](https://github.com/doctolib/yak/compare/v2.22.1...v2.22.2) (2025-07-11)


* **deps:** `MergeMethod*` consts have been split into: `PullRequestMergeMethod*` and `MergeQueueMergeMethod*`.
    - feat!: Add support for pagination options in rules API methods
    ([#&#8203;3562](https://redirect.github.com/google/go-github/issues/3562))
    `GetRulesForBranch`, `GetAllRulesets`, and
    `GetAllRepositoryRulesets` now accept `opts`.
* **PSRE-3650:** support new format for vault-secrets files

### Features

* add terraform provider ([#537](https://github.com/doctolib/yak/issues/537)) ([423e65c](https://github.com/doctolib/yak/commit/423e65cc97fdbedd19142d1b1e938b4fbbadba15))
* **argocd/status:** Add ArgoCD status cmd + expose metrics ([#488](https://github.com/doctolib/yak/issues/488)) ([c908771](https://github.com/doctolib/yak/commit/c9087712f0c9554621249b70ce683e9bc93df891))
* **aurora:** add aurora psql command ([#588](https://github.com/doctolib/yak/issues/588)) ([2b1ddf7](https://github.com/doctolib/yak/commit/2b1ddf733ff99fb43b43a06ef9a3d53e474566c9))
* **aws config generate:** generate config file for AWS access with teleport ([#603](https://github.com/doctolib/yak/issues/603)) ([51e186f](https://github.com/doctolib/yak/commit/51e186f65aea94d44e1db14757b4629e15896bdc))
* **certificates:** Certificates renewal commands ([#420](https://github.com/doctolib/yak/issues/420)) ([697e6a0](https://github.com/doctolib/yak/commit/697e6a0824bb83957975144165f325cc92954b90))
* creates command 'yak secret list-duplicates' ([#579](https://github.com/doctolib/yak/issues/579)) ([390b237](https://github.com/doctolib/yak/commit/390b237c7c4c41ad285fe2180cfbd1c01c4b20ac))
* **EN-1180:** migrates to new GHA runner infra ([#752](https://github.com/doctolib/yak/issues/752)) ([4f2ac6e](https://github.com/doctolib/yak/commit/4f2ac6ee310de1c3849c124170490287442aace0))
* **EN-142:** create & delete aurora clone ([#558](https://github.com/doctolib/yak/issues/558)) ([d1e9761](https://github.com/doctolib/yak/commit/d1e9761bbbeb66208bb98d81333ae42efc60b31c))
* **EN-1501:** Declares cicd-staging Vault ns + cleanup ([#846](https://github.com/doctolib/yak/issues/846)) ([bf88d61](https://github.com/doctolib/yak/commit/bf88d6168c4e0c25d811c058bbd2c7c441d8ffd1))
* **EN-200:** Whitelist philips-labs modules ([#590](https://github.com/doctolib/yak/issues/590)) ([e6784e8](https://github.com/doctolib/yak/commit/e6784e8e0af56e72fb23c8c789e6edc1742a7a4f))
* **en-2219:** Add new repository url for github runner provider ([#920](https://github.com/doctolib/yak/issues/920)) ([6d52f13](https://github.com/doctolib/yak/commit/6d52f130eab90e34cc91e5d88d9a7392c3247457))
* **EN-2932:** Migrate github secrets to vault ([#1153](https://github.com/doctolib/yak/issues/1153)) ([8680606](https://github.com/doctolib/yak/commit/86806062123b5f736e3867f83f2b1a2fbb2e6ef6))
* **EN-337:** use rds client interface to mock client ([#602](https://github.com/doctolib/yak/issues/602)) ([62a1bd4](https://github.com/doctolib/yak/commit/62a1bd48b5542030386f83e5ee84b3e4b684ac00))
* **en-496:** add aurora cluster delete command ([#775](https://github.com/doctolib/yak/issues/775)) ([f98eb9c](https://github.com/doctolib/yak/commit/f98eb9c2d6e9f0b4615116fbd4f908dd8e837e4f))
* **EN-792:** gives ability to specify one DB Cluster parameter group per cluster created from snapshots ([#694](https://github.com/doctolib/yak/issues/694)) ([9af8cf4](https://github.com/doctolib/yak/commit/9af8cf4effcef1ddf9f5b67c89d918b527bef9b2))
* **EN-793:** Rework the way to specify target name for Aurora clones creation ([#678](https://github.com/doctolib/yak/issues/678)) ([21d8d34](https://github.com/doctolib/yak/commit/21d8d3424ca5500c8aef6685b19d4243bfaf5830))
* kafka replication monitoring ([#629](https://github.com/doctolib/yak/issues/629)) ([2db9d03](https://github.com/doctolib/yak/commit/2db9d0368d8b4e22dea86848015941c049a69a92))
* **PSRE-1355:** try to read TFE token from .terraform.d/credentials.tfrc.json ([#373](https://github.com/doctolib/yak/issues/373)) ([3140e25](https://github.com/doctolib/yak/commit/3140e259ef49944f555730b605df112e349c298c))
* **PSRE-2111:** add terraform provider publish workflow ([#669](https://github.com/doctolib/yak/issues/669)) ([b37374b](https://github.com/doctolib/yak/commit/b37374bcaa54b319f90ae947d4a465fdacc83b37))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted in route53 ([#935](https://github.com/doctolib/yak/issues/935)) ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-3127:** Reset workers in kafka ([#704](https://github.com/doctolib/yak/issues/704)) ([6380da8](https://github.com/doctolib/yak/commit/6380da889e34a7a2964b31c6c8fdf258aa6e7a01))
* **PSRE-3650:** prepare refactoring of vault secrets format ([#904](https://github.com/doctolib/yak/issues/904)) ([25d5a62](https://github.com/doctolib/yak/commit/25d5a620bb415feee991ae97014abeeca407cb58))
* **PSRE-3702:** Add yak command to create client and server JWT secrets ([#866](https://github.com/doctolib/yak/issues/866)) ([d8d24a8](https://github.com/doctolib/yak/commit/d8d24a81755b63c0f141c447253631c1e1a88d63))
* **PSRE-3708:** add secret undelete command ([#859](https://github.com/doctolib/yak/issues/859)) ([4097c47](https://github.com/doctolib/yak/commit/4097c47881fb9ff9288c9ebdf5f7f3cef892f5fe))
* **PSRE-3748:** Extend yak secret get command to retrieve secret data matching a provided key ([#869](https://github.com/doctolib/yak/issues/869)) ([6bae36c](https://github.com/doctolib/yak/commit/6bae36c020828ba9aae2feab950e2dc5073af7e1))
* **PSRE-4357:** yak certificate gandi-check to verify your PAT token ([81561f6](https://github.com/doctolib/yak/commit/81561f6643540154b0d5e438c1a19a7341860b83))
* **PSRE-4357:** yak certificate gandi-check to verify your PAT token ([#948](https://github.com/doctolib/yak/issues/948)) ([81561f6](https://github.com/doctolib/yak/commit/81561f6643540154b0d5e438c1a19a7341860b83))
* **PSRE-4388:** Add autogenerated file to allowed provider files ([#1082](https://github.com/doctolib/yak/issues/1082)) ([f6efa16](https://github.com/doctolib/yak/commit/f6efa16497dc413779ebe2e01e7bc0f55d727b52))
* **PSRE-4388:** Add check provider declaration file ([#943](https://github.com/doctolib/yak/issues/943)) ([c577d8a](https://github.com/doctolib/yak/commit/c577d8aeb8c88d9d8fe8b81820ce3ef94ac4c2bb))
* **PSRE-4392:** Add a check for required_providers to be in versions.tf ([#952](https://github.com/doctolib/yak/issues/952)) ([035a3f0](https://github.com/doctolib/yak/commit/035a3f0c06b5a59dfe55103b6504e3b00b3cb0f7))
* **PSRE-4394:** Add allow relative source option ([#1130](https://github.com/doctolib/yak/issues/1130)) ([1c70875](https://github.com/doctolib/yak/commit/1c70875b26e02e4e1b2f42ca37fa638b9011760a))
* **PSRE-4397:** Add cloud declaration in backend.tf check ([#950](https://github.com/doctolib/yak/issues/950)) ([63edb72](https://github.com/doctolib/yak/commit/63edb72cce40b21f38494d56c1cda13d0107b543))
* **PSRE-4454:** get details of x509 certs stored in vault ([bc30c18](https://github.com/doctolib/yak/commit/bc30c18fffec42a0b562aa8bc19b31fe68f77ccf))
* **PSRE-4454:** get details of x509 certs stored in vault ([#947](https://github.com/doctolib/yak/issues/947)) ([bc30c18](https://github.com/doctolib/yak/commit/bc30c18fffec42a0b562aa8bc19b31fe68f77ccf))
* **PSRE-4513:** Build multi arch docker image ([#998](https://github.com/doctolib/yak/issues/998)) ([8c0033b](https://github.com/doctolib/yak/commit/8c0033bbc9f5195d50297c46cd36fc953e33ba1e))
* **PSRE-4513:** Fix yak sanity check ([#1017](https://github.com/doctolib/yak/issues/1017)) ([4af0818](https://github.com/doctolib/yak/commit/4af0818de1ae453de967dcc5aa9cb560b7dbd939))
* **PSRE-4513:** Still trying to fix docker build ([#1003](https://github.com/doctolib/yak/issues/1003)) ([b24171b](https://github.com/doctolib/yak/commit/b24171bd2eeeddb2363e5bf97221d7e805c177ee))
* **PSRE-4517:** Update Dockerfile ([#966](https://github.com/doctolib/yak/issues/966)) ([89c60e8](https://github.com/doctolib/yak/commit/89c60e8d332344c8dd5f3242b8522156fdfba324))
* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([#1066](https://github.com/doctolib/yak/issues/1066)) ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4702:** Add argocd refresh command ([#1160](https://github.com/doctolib/yak/issues/1160)) ([8b7547a](https://github.com/doctolib/yak/commit/8b7547a872f51101f50c3fbf0f7b1643a1d98d0e))
* **PSRE-4702:** Extend yak argocd cli with additional commands ([#1124](https://github.com/doctolib/yak/issues/1124)) ([40f9037](https://github.com/doctolib/yak/commit/40f9037c853b77b0c9a642b6bb1c71f2bf56b8db))
* **PSS-1085:** support getting AWS session on new runners ([#832](https://github.com/doctolib/yak/issues/832)) ([045b35a](https://github.com/doctolib/yak/commit/045b35a29cd25b284ab3136e32230dc1e1dd6b79))
* **SREGREEN-104/argocd:** support for SSO ([#735](https://github.com/doctolib/yak/issues/735)) ([784d2f4](https://github.com/doctolib/yak/commit/784d2f41a1f7029c76929f1165514f798681e9d1))
* **SREGREEN-111/helm:** add external helm charts whitelist check ([#788](https://github.com/doctolib/yak/issues/788)) ([2c33917](https://github.com/doctolib/yak/commit/2c3391773a422c8d530449685899c65e572d5719))
* **SREGREEN-111:** helm external chart whitelist check for charts ([#800](https://github.com/doctolib/yak/issues/800)) ([9e59880](https://github.com/doctolib/yak/commit/9e5988098e7b8737f3637f2b723c9e6f34972dc8))
* **SREGREEN-144:** add helm in docker image ([#786](https://github.com/doctolib/yak/issues/786)) ([0fe9346](https://github.com/doctolib/yak/commit/0fe93462287b520f98ff7aac742a2241c7c8fc4f))
* **SREGREEN-144:** yak helm export ([#737](https://github.com/doctolib/yak/issues/737)) ([f5485e0](https://github.com/doctolib/yak/commit/f5485e014e6db1b39d435d660d2eb1e0a0263f40))
* **SREGREEN-238/helm:** check if we reference prerelease version for internal charts ([#816](https://github.com/doctolib/yak/issues/816)) ([51de454](https://github.com/doctolib/yak/commit/51de4542c6533193af581c449bffece3ac6784b5))
* **SREGREEN-260:** include non-namespaced orphaned resources by default ([#850](https://github.com/doctolib/yak/issues/850)) ([1a296ac](https://github.com/doctolib/yak/commit/1a296ac7477bed5baae8fd3e630e452a3810bc68))
* **SREGREEN-281:** whitelist path where module are fetch from git ([#879](https://github.com/doctolib/yak/issues/879)) ([7023c14](https://github.com/doctolib/yak/commit/7023c144938dcb1f1090893332bc9ac13ab74d13))
* **SREGREEN-338:** Added a set-version command to workspace ([#925](https://github.com/doctolib/yak/issues/925)) ([a9c3873](https://github.com/doctolib/yak/commit/a9c38734e9e759d138a637d75e7b21a51ddc9199))
* **SREGREEN-435:** validate json in secrets ([#940](https://github.com/doctolib/yak/issues/940)) ([2b2cc92](https://github.com/doctolib/yak/commit/2b2cc92067bfc2bc124e741a81a50dfa11684e5f))
* **SREGREEN-443:** include CRDs in the helm templated output ([#922](https://github.com/doctolib/yak/issues/922)) ([9cb3c33](https://github.com/doctolib/yak/commit/9cb3c332d22a94ac67c8789dfa0b394180e40139))
* **SREGREEN-477:** add destroy command ([#930](https://github.com/doctolib/yak/issues/930)) ([c49f2c0](https://github.com/doctolib/yak/commit/c49f2c083c10c72798743097f7fb38adaf16a342))
* **SREGREEN-478:** add clean-vault command ([#951](https://github.com/doctolib/yak/issues/951)) ([b112162](https://github.com/doctolib/yak/commit/b1121629c4fcb5560ad756e94e41a13eb46f2e14))
* **SREGREEN-494:** add create-commit functionality ([#939](https://github.com/doctolib/yak/issues/939)) ([bdfc42b](https://github.com/doctolib/yak/commit/bdfc42ba33fb6f27c4497fa1fb60bf922420f5ea))
* **SREGREEN-543:** add tfeJwtSubjects in logical secrets files ([#1108](https://github.com/doctolib/yak/issues/1108)) ([1e16f7f](https://github.com/doctolib/yak/commit/1e16f7f5be40a95e2c83fe52de7c9999c5c3a89d))
* **SREGREEN-566:** add support for vaultParentNamespace in secret config ([#1053](https://github.com/doctolib/yak/issues/1053)) ([f359ed5](https://github.com/doctolib/yak/commit/f359ed5105e87ca42bb78ffb7978e271b19ef37f))
* **SREGREEN-595:** Added parallelism to couchbase logs collection ([#1064](https://github.com/doctolib/yak/issues/1064)) ([ce4179d](https://github.com/doctolib/yak/commit/ce4179ddfe5a2c39f993fa3aaee95f5d3ca05797))
* **SREGREEN-675:** add support for custom branch name in jira create-branch ([#1118](https://github.com/doctolib/yak/issues/1118)) ([50c1284](https://github.com/doctolib/yak/commit/50c128466788f5418ef953d0333da4af6aca45f1))
* **SREGREEN-675:** jira create-commit to support deriving ID from branch name ([#1119](https://github.com/doctolib/yak/issues/1119)) ([b0032b0](https://github.com/doctolib/yak/commit/b0032b04f9a035a45d289156b0065198fbc556bc))
* **SREGREEN-679:** Add helper to get env var, add YAK_JIRA_PROJECT_KEY ([#1122](https://github.com/doctolib/yak/issues/1122)) ([cf5887c](https://github.com/doctolib/yak/commit/cf5887c31dc05b0673c56bb6b5d458c56a4bb4a0))
* **SREGREEN-69:** display conditions in argocd status ([#896](https://github.com/doctolib/yak/issues/896)) ([acd33ad](https://github.com/doctolib/yak/commit/acd33ad98ebf22ef69c4227fb3b1cdacbfd4ab11))
* **teleport:** add reviewers override option ([#625](https://github.com/doctolib/yak/issues/625)) ([ab38854](https://github.com/doctolib/yak/commit/ab3885452fe82e4b4544bbcc88fba9f0800b6b11))
* **TT-20773:** Extend yak aws aurora psql to allow for running queries against a specific rds instance ([#684](https://github.com/doctolib/yak/issues/684)) ([3d4c7c8](https://github.com/doctolib/yak/commit/3d4c7c85f08dee813058787e78734ddc9b99c2d2))


### Bug Fixes

* add container name in exec pod command ([#448](https://github.com/doctolib/yak/issues/448)) ([372b1c0](https://github.com/doctolib/yak/commit/372b1c00f9945e50d8a83b46526be520218fff6e))
* bump actions/checkout to v4 ([#394](https://github.com/doctolib/yak/issues/394)) ([099980c](https://github.com/doctolib/yak/commit/099980cf00f76938f95fda3cdeb941ad137183d0))
* **certificates:** Adapt after changes in Gandi API ([#511](https://github.com/doctolib/yak/issues/511)) ([eac085b](https://github.com/doctolib/yak/commit/eac085b95ecb2fb08e458f59878b10ca70c37ed2))
* cleaning up of ignored orphaned resources from orphaned resources was messy ([#487](https://github.com/doctolib/yak/issues/487)) ([cba3a37](https://github.com/doctolib/yak/commit/cba3a37b61ba2154454479b72fe0619e90470a16))
* **deps:** update aws-sdk-go-v2 monorepo ([#1088](https://github.com/doctolib/yak/issues/1088)) ([43366ff](https://github.com/doctolib/yak/commit/43366ff0a71288817d93c0d5936a5a236b8364fe))
* **deps:** update aws-sdk-go-v2 monorepo ([#1090](https://github.com/doctolib/yak/issues/1090)) ([3ced281](https://github.com/doctolib/yak/commit/3ced281d0dccb29a53b2a872bc4a421956607fac))
* **deps:** update aws-sdk-go-v2 monorepo ([#1099](https://github.com/doctolib/yak/issues/1099)) ([b5aa3d2](https://github.com/doctolib/yak/commit/b5aa3d25c5d1c00b6ac28a247c338d12ee20c7fe))
* **deps:** update aws-sdk-go-v2 monorepo ([#979](https://github.com/doctolib/yak/issues/979)) ([d5d76b8](https://github.com/doctolib/yak/commit/d5d76b8df9698528d6c57db33ea260de36373642))
* **deps:** update dependency go to v1.24.4 ([#1127](https://github.com/doctolib/yak/issues/1127)) ([a37ac8d](https://github.com/doctolib/yak/commit/a37ac8d6d3d7302d1faab3b2cb9731bb0882f2fc))
* **deps:** update github.com/hashicorp/terraform-config-inspect digest to d2d12f9 ([#971](https://github.com/doctolib/yak/issues/971)) ([0a002bf](https://github.com/doctolib/yak/commit/0a002bfe33dcf49549d8081557e17f4c1cb4cafe))
* **deps:** update github.com/hashicorp/terraform-config-inspect digest to f4c50e6 ([#1038](https://github.com/doctolib/yak/issues/1038)) ([7fd9ecb](https://github.com/doctolib/yak/commit/7fd9ecbc78976f18416f7d8d317d94d56e29c37a))
* **deps:** update kubernetes packages to v0.33.2 ([#980](https://github.com/doctolib/yak/issues/980)) ([a9e89fb](https://github.com/doctolib/yak/commit/a9e89fbd0b308e0ce40dc560d4114b876980dd98))
* **deps:** update module github.com/argoproj/argo-cd/v2 to v2.13.8 [security] ([#961](https://github.com/doctolib/yak/issues/961)) ([db0e5bc](https://github.com/doctolib/yak/commit/db0e5bca5eda8908515f8b17ebbc495efda59913))
* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.3 ([#973](https://github.com/doctolib/yak/issues/973)) ([86134f1](https://github.com/doctolib/yak/commit/86134f1919bcec3165c553d922817d7327ce0fb9))
* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.4 ([#1051](https://github.com/doctolib/yak/issues/1051)) ([e1d4b1a](https://github.com/doctolib/yak/commit/e1d4b1ab58feeeb61ea086932f4dd4077e27649a))
* **deps:** update module github.com/aws/aws-sdk-go to v1.55.7 ([#981](https://github.com/doctolib/yak/issues/981)) ([5d54f61](https://github.com/doctolib/yak/commit/5d54f616d39f1533b810364dc218811cc5f8e94f))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/ecr to v1.45.0 ([#1097](https://github.com/doctolib/yak/issues/1097)) ([e724949](https://github.com/doctolib/yak/commit/e724949c193afde9d5499dc0714a2b7e33e23b3e))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.96.0 ([#1054](https://github.com/doctolib/yak/issues/1054)) ([3f277d6](https://github.com/doctolib/yak/commit/3f277d6b8d23b3e738734e5f8b6c577be4ed6ddc))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.97.2 ([#1091](https://github.com/doctolib/yak/issues/1091)) ([2dc05b2](https://github.com/doctolib/yak/commit/2dc05b2aa42470ef1ee4f4a43e86b5a8f44b93c1))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.98.0 ([#1125](https://github.com/doctolib/yak/issues/1125)) ([628aa0f](https://github.com/doctolib/yak/commit/628aa0ff1e920c6c14dc6d1437a86ba01301c4e2))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.99.0 ([#1146](https://github.com/doctolib/yak/issues/1146)) ([a525f03](https://github.com/doctolib/yak/commit/a525f0301d6363414003ad395fb3e1949ca5c488))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.99.1 ([#1149](https://github.com/doctolib/yak/issues/1149)) ([0445bfb](https://github.com/doctolib/yak/commit/0445bfbc9852656478c6aaee631b19286968a528))
* **deps:** update module github.com/aws/smithy-go to v1.22.3 ([#982](https://github.com/doctolib/yak/issues/982)) ([d0b39da](https://github.com/doctolib/yak/commit/d0b39da9f0879015d174b14a759521e2bc212f67))
* **deps:** update module github.com/aws/smithy-go to v1.22.4 ([#1095](https://github.com/doctolib/yak/issues/1095)) ([082fbb6](https://github.com/doctolib/yak/commit/082fbb65bbcaa83eec58ab5e334dc0fc28d1e72e))
* **deps:** update module github.com/birdayz/kaf to v0.2.13 ([#974](https://github.com/doctolib/yak/issues/974)) ([0f75527](https://github.com/doctolib/yak/commit/0f755279f4d54b4533d8dc86f3356849d2c9daab))
* **deps:** update module github.com/coreos/go-oidc/v3 to v3.14.1 ([#983](https://github.com/doctolib/yak/issues/983)) ([bdc5b1a](https://github.com/doctolib/yak/commit/bdc5b1ab6d24bf085b90465e26614362c36f0cf4))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.37.1 ([#984](https://github.com/doctolib/yak/issues/984)) ([e209e14](https://github.com/doctolib/yak/commit/e209e1446f631a96756dba0a981103ea2bc480e0))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.38.0 ([#1069](https://github.com/doctolib/yak/issues/1069)) ([aa76205](https://github.com/doctolib/yak/commit/aa762053a2d568ee9b5731e4461b09b327d07746))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.39.0 ([#1098](https://github.com/doctolib/yak/issues/1098)) ([e9fbb61](https://github.com/doctolib/yak/commit/e9fbb61b07400e3c590e776a3eb4ffe629cf14b2))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.40.0 ([#1121](https://github.com/doctolib/yak/issues/1121)) ([9935480](https://github.com/doctolib/yak/commit/9935480492746b652d05839aaaa1291f6229c296))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.41.0 ([#1126](https://github.com/doctolib/yak/issues/1126)) ([1a64067](https://github.com/doctolib/yak/commit/1a6406745e8492894d2264a3906a9eaf4bd15317))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.42.0 ([#1148](https://github.com/doctolib/yak/issues/1148)) ([40630f9](https://github.com/doctolib/yak/commit/40630f9940edc66ee7cf5a59c33e7dec44bb43ac))
* **deps:** update module github.com/fatih/color to v1.18.0 ([#985](https://github.com/doctolib/yak/issues/985)) ([931bf81](https://github.com/doctolib/yak/commit/931bf8101f30923a9987e91e1a694cd39b8123b2))
* **deps:** update module github.com/go-git/go-git/v5 to v5.13.0 [security] ([#962](https://github.com/doctolib/yak/issues/962)) ([7558f5b](https://github.com/doctolib/yak/commit/7558f5bf9f1a3277396f949577b2c4fb2bf2cc9f))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.0 ([#986](https://github.com/doctolib/yak/issues/986)) ([203a42b](https://github.com/doctolib/yak/commit/203a42beddecc6f5541323b31b6b06f34c21c7f1))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.1 ([#1077](https://github.com/doctolib/yak/issues/1077)) ([8364f4c](https://github.com/doctolib/yak/commit/8364f4ccc8d96b39b6361bd73fd5661e5955bf64))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.2 ([#1089](https://github.com/doctolib/yak/issues/1089)) ([1e3e5cc](https://github.com/doctolib/yak/commit/1e3e5cc4322f63c55585b0cdcad5daeabf400f46))
* **deps:** update module github.com/golang-jwt/jwt/v4 to v4.5.2 [security] ([#960](https://github.com/doctolib/yak/issues/960)) ([09c8038](https://github.com/doctolib/yak/commit/09c8038e45de2543fc525898bb6fd2592de1ac1f))
* **deps:** update module github.com/golang-jwt/jwt/v4 to v5 ([#1044](https://github.com/doctolib/yak/issues/1044)) ([9ce844c](https://github.com/doctolib/yak/commit/9ce844c8bf99a78e1944f60ca0b4f511cdaa50e9))
* **deps:** update module github.com/google/go-github/v52 to v72 ([#1045](https://github.com/doctolib/yak/issues/1045)) ([e04cbb0](https://github.com/doctolib/yak/commit/e04cbb0cd8146bd925420ba5be7314719b82c4b0))
* **deps:** update module github.com/google/go-github/v72 to v73 ([#1145](https://github.com/doctolib/yak/issues/1145)) ([d523ffa](https://github.com/doctolib/yak/commit/d523ffa1341edd1570ee131be5b1d20481b9792f))
* **deps:** update module github.com/google/go-github/v72 to v73 ([#1152](https://github.com/doctolib/yak/issues/1152)) ([90d9a26](https://github.com/doctolib/yak/commit/90d9a268f1f840c1d0f18f4dbadf0612c4325e23))
* **deps:** update module github.com/hashicorp/go-tfe to v1.80.0 ([#987](https://github.com/doctolib/yak/issues/987)) ([c77a99e](https://github.com/doctolib/yak/commit/c77a99eba20280d4962438114b9e4118b9f5eb4e))
* **deps:** update module github.com/hashicorp/go-tfe to v1.81.0 ([#1058](https://github.com/doctolib/yak/issues/1058)) ([38f564c](https://github.com/doctolib/yak/commit/38f564c4bafc4f3e5342af8f3d4e29a6a8330bc1))
* **deps:** update module github.com/hashicorp/go-tfe to v1.82.0 ([#1092](https://github.com/doctolib/yak/issues/1092)) ([1631036](https://github.com/doctolib/yak/commit/16310365f6139c3781d76097684dd1fa9f3315af))
* **deps:** update module github.com/hashicorp/go-tfe to v1.83.0 ([#1100](https://github.com/doctolib/yak/issues/1100)) ([ec66302](https://github.com/doctolib/yak/commit/ec6630206980bca6140f0cb2b01e0714dce3ac18))
* **deps:** update module github.com/hashicorp/go-tfe to v1.84.0 ([#1114](https://github.com/doctolib/yak/issues/1114)) ([8407de3](https://github.com/doctolib/yak/commit/8407de3921acb9a3f01ff668b54c30882e460de2))
* **deps:** update module github.com/hashicorp/go-tfe to v1.85.0 ([#1150](https://github.com/doctolib/yak/issues/1150)) ([39d071f](https://github.com/doctolib/yak/commit/39d071fa399455b3e51bd95749fa52038caaede3))
* **deps:** update module github.com/hashicorp/hcl/v2 to v2.23.0 ([#989](https://github.com/doctolib/yak/issues/989)) ([c47a6a5](https://github.com/doctolib/yak/commit/c47a6a5c06a82c8328e23166035a4ba7e3f4a8dc))
* **deps:** update module github.com/hashicorp/hcl/v2 to v2.24.0 ([#1154](https://github.com/doctolib/yak/issues/1154)) ([0d578bf](https://github.com/doctolib/yak/commit/0d578bfcfa3a66de8fd0f08f3cf181ce94cfd089))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.14.1 ([#991](https://github.com/doctolib/yak/issues/991)) ([e0af416](https://github.com/doctolib/yak/commit/e0af4167f5d5b6ab92b4e740752b961760d143a8))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.15.0 ([#1049](https://github.com/doctolib/yak/issues/1049)) ([9feb107](https://github.com/doctolib/yak/commit/9feb1071b2d539d43524d741c44f4a972220be98))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework-jsontypes to v0.2.0 ([#992](https://github.com/doctolib/yak/issues/992)) ([015549d](https://github.com/doctolib/yak/commit/015549d3f90e934a4e2d8c456187950e262b6a69))
* **deps:** update module github.com/hashicorp/vault to v1.19.3 [security] ([#963](https://github.com/doctolib/yak/issues/963)) ([c7edbdc](https://github.com/doctolib/yak/commit/c7edbdc5af963b9492ac13da72024b5ac1404074))
* **deps:** update module github.com/hashicorp/vault to v1.19.5 ([#1112](https://github.com/doctolib/yak/issues/1112)) ([fbff5fa](https://github.com/doctolib/yak/commit/fbff5faaa1390387ec25048482b18714ffe4255f))
* **deps:** update module github.com/hashicorp/vault-plugin-secrets-kv to v0.24.1 ([#993](https://github.com/doctolib/yak/issues/993)) ([e396d19](https://github.com/doctolib/yak/commit/e396d19fdab3dcd00004d9924e21ebcff6119c9e))
* **deps:** update module github.com/hashicorp/vault/api to v1.16.0 ([#994](https://github.com/doctolib/yak/issues/994)) ([13772d9](https://github.com/doctolib/yak/commit/13772d9855d9e6f6694f5d75734514f552b7c91a))
* **deps:** update module github.com/hashicorp/vault/api to v1.20.0 ([#1078](https://github.com/doctolib/yak/issues/1078)) ([ce75d2f](https://github.com/doctolib/yak/commit/ce75d2f18e5c945fc13e2f12f19e3af11b977e09))
* **deps:** update module github.com/hashicorp/vault/sdk to v0.18.0 ([#1115](https://github.com/doctolib/yak/issues/1115)) ([3b93fd4](https://github.com/doctolib/yak/commit/3b93fd4363990c546625f1c5ec7f124efa5cd72f))
* **deps:** update module github.com/ibm/sarama to v1.45.1 ([#995](https://github.com/doctolib/yak/issues/995)) ([daf9404](https://github.com/doctolib/yak/commit/daf9404efc58112ca225b8dc288208409cb6144e))
* **deps:** update module github.com/ibm/sarama to v1.45.2 ([#1068](https://github.com/doctolib/yak/issues/1068)) ([1faef9b](https://github.com/doctolib/yak/commit/1faef9bfad695a4e687322bc01b15b50804f6a0b))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([#996](https://github.com/doctolib/yak/issues/996)) ([b556921](https://github.com/doctolib/yak/commit/b55692101288c4ce19c191d589f4da018dd04f37))
* **deps:** update module github.com/prometheus/common to v0.63.0 ([#997](https://github.com/doctolib/yak/issues/997)) ([5e22148](https://github.com/doctolib/yak/commit/5e22148fb8e8c42e811fd6ee5f7b4c9f102914b8))
* **deps:** update module github.com/prometheus/common to v0.64.0 ([#1041](https://github.com/doctolib/yak/issues/1041)) ([8d69e14](https://github.com/doctolib/yak/commit/8d69e14225313a7fd38217a682705be7c76d680a))
* **deps:** update module github.com/prometheus/common to v0.65.0 ([#1117](https://github.com/doctolib/yak/issues/1117)) ([13add5f](https://github.com/doctolib/yak/commit/13add5fcbb2cfc97ecf78283241f2661df8c3b87))
* **deps:** update module github.com/schollz/progressbar/v3 to v3.18.0 ([#1006](https://github.com/doctolib/yak/issues/1006)) ([2b60e89](https://github.com/doctolib/yak/commit/2b60e8905f198779730c2edc6fb286a9685b33ca))
* **deps:** update module github.com/spf13/cobra to v1.9.1 ([#1015](https://github.com/doctolib/yak/issues/1015)) ([96490ee](https://github.com/doctolib/yak/commit/96490ee9769c591fb813a0325815bdeb41546cf3))
* **deps:** update module github.com/spf13/viper to v1.20.1 ([#1025](https://github.com/doctolib/yak/issues/1025)) ([4594dba](https://github.com/doctolib/yak/commit/4594dba7e13c0b43f2b3bb32a417edb7d754a850))
* **deps:** update module github.com/zalando/go-keyring to v0.2.6 ([#975](https://github.com/doctolib/yak/issues/975)) ([5ef0beb](https://github.com/doctolib/yak/commit/5ef0beb1e2ff092a1d55812efe9535568392bacf))
* **deps:** update module github.com/zclconf/go-cty to v1.16.2 ([#1027](https://github.com/doctolib/yak/issues/1027)) ([0aa93db](https://github.com/doctolib/yak/commit/0aa93db625c987c794ffcc9b2fdadd535a304515))
* **deps:** update module github.com/zclconf/go-cty to v1.16.3 ([#1048](https://github.com/doctolib/yak/issues/1048)) ([639e906](https://github.com/doctolib/yak/commit/639e90656b86a9c03a85e04dd10ce66daffbe3ac))
* **deps:** update module golang.org/x/crypto to v0.35.0 [security] ([#964](https://github.com/doctolib/yak/issues/964)) ([5719d16](https://github.com/doctolib/yak/commit/5719d1651bb0ab9e4cc5d1a0e396647340c91e44))
* **deps:** update module golang.org/x/crypto to v0.38.0 ([#1028](https://github.com/doctolib/yak/issues/1028)) ([7399ae2](https://github.com/doctolib/yak/commit/7399ae2b60992e93ed2c7a0bd070bdbdc4da87f0))
* **deps:** update module golang.org/x/crypto to v0.39.0 ([#1084](https://github.com/doctolib/yak/issues/1084)) ([ef0937f](https://github.com/doctolib/yak/commit/ef0937fe8536cc3bd0333fb9960f80958039da28))
* **deps:** update module golang.org/x/crypto to v0.40.0 ([#1161](https://github.com/doctolib/yak/issues/1161)) ([4b2d7d2](https://github.com/doctolib/yak/commit/4b2d7d2d1a0373d105bcbd662d2528b970db878f))
* **deps:** update module golang.org/x/mod to v0.24.0 ([#1029](https://github.com/doctolib/yak/issues/1029)) ([f46e796](https://github.com/doctolib/yak/commit/f46e79660786edac352397435ec682d429dc8ead))
* **deps:** update module golang.org/x/mod to v0.26.0 ([#1157](https://github.com/doctolib/yak/issues/1157)) ([58ee469](https://github.com/doctolib/yak/commit/58ee4693530ec388074920fd06517ee74567d04e))
* **deps:** update module golang.org/x/oauth2 to v0.30.0 ([#1030](https://github.com/doctolib/yak/issues/1030)) ([17e98c1](https://github.com/doctolib/yak/commit/17e98c1507440171614091a9dfd486611714c4f0))
* **deps:** update module golang.org/x/sync to v0.14.0 ([#1031](https://github.com/doctolib/yak/issues/1031)) ([cc3a744](https://github.com/doctolib/yak/commit/cc3a744cf2b0de119154e9640c42304c7b2e5c2a))
* **deps:** update module golang.org/x/sync to v0.15.0 ([#1086](https://github.com/doctolib/yak/issues/1086)) ([db186a6](https://github.com/doctolib/yak/commit/db186a673da677cde92c832b49a97c35b7f4396f))
* **deps:** update module golang.org/x/sync to v0.16.0 ([#1158](https://github.com/doctolib/yak/issues/1158)) ([580e1a2](https://github.com/doctolib/yak/commit/580e1a2432b05623a839e67799a80c3065091ed7))
* **deps:** update module golang.org/x/term to v0.32.0 ([#1033](https://github.com/doctolib/yak/issues/1033)) ([7ee03fc](https://github.com/doctolib/yak/commit/7ee03fcede537968d4d2fe8a7611ff1a2a2c938c))
* **deps:** update module golang.org/x/term to v0.33.0 ([#1159](https://github.com/doctolib/yak/issues/1159)) ([83ea9ef](https://github.com/doctolib/yak/commit/83ea9ef618fad2967db4ce4239bb3fa532337297))
* **deps:** update module sigs.k8s.io/yaml to v1.5.0 ([#1132](https://github.com/doctolib/yak/issues/1132)) ([fa0d991](https://github.com/doctolib/yak/commit/fa0d9913e90f9c07890d7b2f56a2728d5b737f56))
* **en-1244:** add workflow permissions for promote ci version workflow ([#783](https://github.com/doctolib/yak/issues/783)) ([53478d2](https://github.com/doctolib/yak/commit/53478d286b903d2f7b16cd8557a9047e4e1e0e82))
* **en-1244:** remove space ([#779](https://github.com/doctolib/yak/issues/779)) ([87956fc](https://github.com/doctolib/yak/commit/87956fc2dc57c4cf249c483faa06ce2109fc21a3))
* **EN-142:** Ensures Aurora clone uses same SecurityGroups as source ([#626](https://github.com/doctolib/yak/issues/626)) ([0140ce8](https://github.com/doctolib/yak/commit/0140ce854aaa3355fda000bb8053b36e1a00fd54))
* **EN-2538:** fixes ordering of aurora clone source/target args ([#941](https://github.com/doctolib/yak/issues/941)) ([4be2c16](https://github.com/doctolib/yak/commit/4be2c16196e11b46b2f9e479cbf47f8b939fdd19))
* in-cluster kube config ([#501](https://github.com/doctolib/yak/issues/501)) ([0ba5c35](https://github.com/doctolib/yak/commit/0ba5c35614c09c8f1265774142ec0a03a6957e4b))
* issues found by linter ([#361](https://github.com/doctolib/yak/issues/361)) ([ad29686](https://github.com/doctolib/yak/commit/ad296863b1c7b6ade6d763b65d8244d68f12d7a8))
* **kube secret check:** unexpected error after common namespace split ([#513](https://github.com/doctolib/yak/issues/513)) ([8e105bc](https://github.com/doctolib/yak/commit/8e105bc3bea6561bc5d714481be9ecffdb14d5dc))
* **provider check:** allow terraform_data resource ([#577](https://github.com/doctolib/yak/issues/577)) ([4c53f26](https://github.com/doctolib/yak/commit/4c53f262bcc62947fbe9a9f4d748d211f9bff63a))
* **PSRE-1569:** fix check command after split ([#412](https://github.com/doctolib/yak/issues/412)) ([3577c3d](https://github.com/doctolib/yak/commit/3577c3d2e876dcfa5c47f2b462b46234052426e2))
* **PSRE-2088:** add ruby to support ruby helm post-render ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-2088:** add ruby to support ruby helm post-render ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-2088:** add ruby to support ruby helm post-render ([#931](https://github.com/doctolib/yak/issues/931)) ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-3054:** use a dedicated clone for each operation ([#695](https://github.com/doctolib/yak/issues/695)) ([a6d17e3](https://github.com/doctolib/yak/commit/a6d17e38a342dfd06a8a4de63b15fd950a0763db))
* **PSRE-3748:** Fix returned secret when getting secret data with keys matching provided key ([#871](https://github.com/doctolib/yak/issues/871)) ([354c477](https://github.com/doctolib/yak/commit/354c477fb1d76b2f049af2b843e5a70e2bceacd2))
* **PSRE-3784:** Yak secret jwt server not updating CI secret ([#878](https://github.com/doctolib/yak/issues/878)) ([1adeab5](https://github.com/doctolib/yak/commit/1adeab508084f849577cf61a91435598d5955a8a))
* **PSRE-4018:** Update JWT token creation to use snake case service name in key for JWT token services ([#916](https://github.com/doctolib/yak/issues/916)) ([56ea27f](https://github.com/doctolib/yak/commit/56ea27f802ee777acb107d164f828ded31f69ed8))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([#949](https://github.com/doctolib/yak/issues/949)) ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4388:** Fix error count and tests ([#1034](https://github.com/doctolib/yak/issues/1034)) ([a1d6a79](https://github.com/doctolib/yak/commit/a1d6a7997415e7572903c9a959d6c0e056b7214c))
* **PSRE-4513:** Fix build of ARM image ([#1001](https://github.com/doctolib/yak/issues/1001)) ([74bc8ed](https://github.com/doctolib/yak/commit/74bc8ed956f9789f08fad69f956b536a77286ac5))
* **PSRE-4517:** fix and refactor some provider and declaration checks ([#955](https://github.com/doctolib/yak/issues/955)) ([968ebc0](https://github.com/doctolib/yak/commit/968ebc0d4c8d17fb8f4329fa221d6ae708b7b683))
* **PSRE-4517:** Update Dockerfile ([#968](https://github.com/doctolib/yak/issues/968)) ([fe8502f](https://github.com/doctolib/yak/commit/fe8502ff7fb9449001d754167161b04f6865abda))
* **PSRE-4702:** Fix for dashboard command when specifying app ([#1133](https://github.com/doctolib/yak/issues/1133)) ([0c34044](https://github.com/doctolib/yak/commit/0c3404423b14251f7f70a6f55c2637bec5001359))
* **PSRE-4702:** Refactor  argocd diff command ([#1129](https://github.com/doctolib/yak/issues/1129)) ([f430d6f](https://github.com/doctolib/yak/commit/f430d6f08cb5f743232a97488dec38f719d74345))
* **PSS-914:** Yak argocd commands failed when KUBECONFIG var is too long ([#721](https://github.com/doctolib/yak/issues/721)) ([055e79b](https://github.com/doctolib/yak/commit/055e79b59ccfab9fc4e66fca525bbc0db211e2e4))
* rename maintainer team ([#457](https://github.com/doctolib/yak/issues/457)) ([80ab7e4](https://github.com/doctolib/yak/commit/80ab7e4dd9b417fc2df40ae17de320bb00351572))
* **secret:** terraform path fed by yak provider is not accessible ([#559](https://github.com/doctolib/yak/issues/559)) ([d811b5b](https://github.com/doctolib/yak/commit/d811b5b40f81d71c42c9d5124e5f46b97cf1d71d))
* **SREBLUE-001:** Fixes for jira and github commands ([#645](https://github.com/doctolib/yak/issues/645)) ([1eb2739](https://github.com/doctolib/yak/commit/1eb273929b80c4440e75b0c1da6cf2e04fabaaa6))
* **SREGREEN-111:** check for chart is incomplete ([#801](https://github.com/doctolib/yak/issues/801)) ([488aa09](https://github.com/doctolib/yak/commit/488aa09b05545d6ee2f2bbc4f5d1dd506193c412))
* **SREGREEN-111:** incomplete error message ([#805](https://github.com/doctolib/yak/issues/805)) ([b868fe3](https://github.com/doctolib/yak/commit/b868fe30b4ad8c092e70975bf2c5418e2bd626ce))
* **SREGREEN-111:** panic when there is no dependency ([#804](https://github.com/doctolib/yak/issues/804)) ([54b85f0](https://github.com/doctolib/yak/commit/54b85f03c715ac69c1aa6cfd2e4b878ed7b16174))
* **SREGREEN-144:** fix config bug and add tests ([#784](https://github.com/doctolib/yak/issues/784)) ([62a7d5e](https://github.com/doctolib/yak/commit/62a7d5eaf5789c63d382ad15cb302a7937ef52a3))
* **SREGREEN-144:** fix managed files ([#787](https://github.com/doctolib/yak/issues/787)) ([82f75b8](https://github.com/doctolib/yak/commit/82f75b83dd4f7db5bd24f0853afc08ef415662ee))
* **SREGREEN-215:** add bybasstsh flag deleted by mistake ([#778](https://github.com/doctolib/yak/issues/778)) ([f4328e3](https://github.com/doctolib/yak/commit/f4328e3fd33e73e3465c5522cc38f318db16a91c))
* **SREGREEN-221:** helm template printing logs to stdout... ([#790](https://github.com/doctolib/yak/issues/790)) ([a50bdb3](https://github.com/doctolib/yak/commit/a50bdb342ea3a14a7e43b2e2301057a1df29de66))
* **SREGREEN-228:** lags calculation ([#820](https://github.com/doctolib/yak/issues/820)) ([1206ea0](https://github.com/doctolib/yak/commit/1206ea0689a943f4b08b652b177aed8bb86db67a))
* **SREGREEN-350:** automate common semantic release flow ([#953](https://github.com/doctolib/yak/issues/953)) ([cde76a9](https://github.com/doctolib/yak/commit/cde76a9f6111c86165bce1b1e3bbb23914717e3e))
* **SREGREEN-350:** configure renovate to run `go mod tidy` ([#1050](https://github.com/doctolib/yak/issues/1050)) ([bdb1989](https://github.com/doctolib/yak/commit/bdb198905c9b6dc4f8a3c4c7c42c154ecf67920e))
* **SREGREEN-350:** use beefier runner for publish workflow ([#959](https://github.com/doctolib/yak/issues/959)) ([85236c0](https://github.com/doctolib/yak/commit/85236c0526b482a3f9bc8b2ba610eba3da4ff232))
* **SREGREEN-449:** --config flag not working ([#923](https://github.com/doctolib/yak/issues/923)) ([2506b17](https://github.com/doctolib/yak/commit/2506b17bdd5f1af482af13911e2e3ff7b0d25fb8))
* **SREGREEN-48:** update argocd cli to use "main" project ([#641](https://github.com/doctolib/yak/issues/641)) ([d0e134f](https://github.com/doctolib/yak/commit/d0e134f6ac1ce385c3554809e94ab6bf7e6af750))
* **SREGREEN-499:** create_branch: do not clean local index + hide completed tasks ([#944](https://github.com/doctolib/yak/issues/944)) ([c17ce75](https://github.com/doctolib/yak/commit/c17ce7557455e35554b8a14ead7bde611a841380))
* **SREGREEN-575:** honor the -a option in 'yak argocd status' ([#1062](https://github.com/doctolib/yak/issues/1062)) ([dda7cb0](https://github.com/doctolib/yak/commit/dda7cb0ea7c95042028e2209b7c5b0545153b6f1))
* **SREGREEN-57:** sort alphabetically output of yak argocd status ([#847](https://github.com/doctolib/yak/issues/847)) ([d8a5b06](https://github.com/doctolib/yak/commit/d8a5b06a939de10ec7fc5af07cf487006d144609))
* **SREGREEN-635:** argocd suspend to support UI suspensions ([#1072](https://github.com/doctolib/yak/issues/1072)) ([e0d2fc0](https://github.com/doctolib/yak/commit/e0d2fc0fc3201a2a0c3c0b7c18cbe0185967afd4))
* **SREGREEN-640/jwt:** config flag not being honored ([#1080](https://github.com/doctolib/yak/issues/1080)) ([57c6512](https://github.com/doctolib/yak/commit/57c6512a792b395c70e77e2e2465982f77fa4cc7))
* **SREGREEN-70/argocd:** fix rendering of argocd status ([#761](https://github.com/doctolib/yak/issues/761)) ([4eca3e0](https://github.com/doctolib/yak/commit/4eca3e063cf3d4353893bd057c9f1d25f278206d))
* **SREGREEN-751:** make jira status check case insensitive ([#1163](https://github.com/doctolib/yak/issues/1163)) ([07f60f3](https://github.com/doctolib/yak/commit/07f60f3fa16bfbd08a48ff1264c498a9b8fdc7b3))
* use logrus lib instead of default lib ([#483](https://github.com/doctolib/yak/issues/483)) ([f3b3bbf](https://github.com/doctolib/yak/commit/f3b3bbf29bd4e1d39132d6f66c0285f1ff531c33))


### Reverts

* **EN-1035:** revert change on publish.yml ([#733](https://github.com/doctolib/yak/issues/733)) ([5fbf1a3](https://github.com/doctolib/yak/commit/5fbf1a370b5f42b98ff2e11d239f448022309a8a))


### Code Refactoring

* **PSRE-3650:** support new format for vault-secrets files ([#902](https://github.com/doctolib/yak/issues/902)) ([856084a](https://github.com/doctolib/yak/commit/856084a5aef6156af4fffe7e6050375f8bf5a5eb))

## [2.22.1](https://github.com/doctolib/yak/compare/v2.22.0...v2.22.1) (2025-07-11)


### Bug Fixes

* **SREGREEN-751:** make jira status check case insensitive ([#1163](https://github.com/doctolib/yak/issues/1163)) ([07f60f3](https://github.com/doctolib/yak/commit/07f60f3fa16bfbd08a48ff1264c498a9b8fdc7b3))

## [2.22.0](https://github.com/doctolib/yak/compare/v2.21.4...v2.22.0) (2025-07-11)


### Features

* **EN-2932:** Migrate github secrets to vault ([#1153](https://github.com/doctolib/yak/issues/1153)) ([8680606](https://github.com/doctolib/yak/commit/86806062123b5f736e3867f83f2b1a2fbb2e6ef6))
* **PSRE-4394:** Add allow relative source option ([#1130](https://github.com/doctolib/yak/issues/1130)) ([1c70875](https://github.com/doctolib/yak/commit/1c70875b26e02e4e1b2f42ca37fa638b9011760a))
* **PSRE-4702:** Add argocd refresh command ([#1160](https://github.com/doctolib/yak/issues/1160)) ([8b7547a](https://github.com/doctolib/yak/commit/8b7547a872f51101f50c3fbf0f7b1643a1d98d0e))


### Bug Fixes

* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.99.0 ([#1146](https://github.com/doctolib/yak/issues/1146)) ([a525f03](https://github.com/doctolib/yak/commit/a525f0301d6363414003ad395fb3e1949ca5c488))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.99.1 ([#1149](https://github.com/doctolib/yak/issues/1149)) ([0445bfb](https://github.com/doctolib/yak/commit/0445bfbc9852656478c6aaee631b19286968a528))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.42.0 ([#1148](https://github.com/doctolib/yak/issues/1148)) ([40630f9](https://github.com/doctolib/yak/commit/40630f9940edc66ee7cf5a59c33e7dec44bb43ac))
* **deps:** update module github.com/google/go-github/v72 to v73 ([#1145](https://github.com/doctolib/yak/issues/1145)) ([d523ffa](https://github.com/doctolib/yak/commit/d523ffa1341edd1570ee131be5b1d20481b9792f))
* **deps:** update module github.com/google/go-github/v72 to v73 ([#1152](https://github.com/doctolib/yak/issues/1152)) ([90d9a26](https://github.com/doctolib/yak/commit/90d9a268f1f840c1d0f18f4dbadf0612c4325e23))
* **deps:** update module github.com/hashicorp/go-tfe to v1.85.0 ([#1150](https://github.com/doctolib/yak/issues/1150)) ([39d071f](https://github.com/doctolib/yak/commit/39d071fa399455b3e51bd95749fa52038caaede3))
* **deps:** update module github.com/hashicorp/hcl/v2 to v2.24.0 ([#1154](https://github.com/doctolib/yak/issues/1154)) ([0d578bf](https://github.com/doctolib/yak/commit/0d578bfcfa3a66de8fd0f08f3cf181ce94cfd089))
* **deps:** update module golang.org/x/crypto to v0.40.0 ([#1161](https://github.com/doctolib/yak/issues/1161)) ([4b2d7d2](https://github.com/doctolib/yak/commit/4b2d7d2d1a0373d105bcbd662d2528b970db878f))
* **deps:** update module golang.org/x/mod to v0.26.0 ([#1157](https://github.com/doctolib/yak/issues/1157)) ([58ee469](https://github.com/doctolib/yak/commit/58ee4693530ec388074920fd06517ee74567d04e))
* **deps:** update module golang.org/x/sync to v0.16.0 ([#1158](https://github.com/doctolib/yak/issues/1158)) ([580e1a2](https://github.com/doctolib/yak/commit/580e1a2432b05623a839e67799a80c3065091ed7))
* **deps:** update module golang.org/x/term to v0.33.0 ([#1159](https://github.com/doctolib/yak/issues/1159)) ([83ea9ef](https://github.com/doctolib/yak/commit/83ea9ef618fad2967db4ce4239bb3fa532337297))

## [2.21.4](https://github.com/doctolib/yak/compare/v2.21.3...v2.21.4) (2025-06-26)


### Changes

* **deps:** `MergeMethod*` consts have been split into: `PullRequestMergeMethod*` and `MergeQueueMergeMethod*`.
    - feat!: Add support for pagination options in rules API methods
    ([#&#8203;3562](https://redirect.github.com/google/go-github/issues/3562))
    `GetRulesForBranch`, `GetAllRulesets`, and
    `GetAllRepositoryRulesets` now accept `opts`.
* **PSRE-3650:** support new format for vault-secrets files

### Features

* add terraform module dependency command ([#316](https://github.com/doctolib/yak/issues/316)) ([7f6dac9](https://github.com/doctolib/yak/commit/7f6dac948fd8c0a0804892708870d80d406b8213))
* add terraform provider ([#537](https://github.com/doctolib/yak/issues/537)) ([423e65c](https://github.com/doctolib/yak/commit/423e65cc97fdbedd19142d1b1e938b4fbbadba15))
* **argocd/status:** Add ArgoCD status cmd + expose metrics ([#488](https://github.com/doctolib/yak/issues/488)) ([c908771](https://github.com/doctolib/yak/commit/c9087712f0c9554621249b70ce683e9bc93df891))
* **aurora:** add aurora psql command ([#588](https://github.com/doctolib/yak/issues/588)) ([2b1ddf7](https://github.com/doctolib/yak/commit/2b1ddf733ff99fb43b43a06ef9a3d53e474566c9))
* **aws config generate:** generate config file for AWS access with teleport ([#603](https://github.com/doctolib/yak/issues/603)) ([51e186f](https://github.com/doctolib/yak/commit/51e186f65aea94d44e1db14757b4629e15896bdc))
* **bump:** allow to specify branch name used to create pull-request ([#314](https://github.com/doctolib/yak/issues/314)) ([6c4f064](https://github.com/doctolib/yak/commit/6c4f064cddbf9fb4f8da69696bac9b26e76f082d))
* **certificates:** Certificates renewal commands ([#420](https://github.com/doctolib/yak/issues/420)) ([697e6a0](https://github.com/doctolib/yak/commit/697e6a0824bb83957975144165f325cc92954b90))
* creates command 'yak secret list-duplicates' ([#579](https://github.com/doctolib/yak/issues/579)) ([390b237](https://github.com/doctolib/yak/commit/390b237c7c4c41ad285fe2180cfbd1c01c4b20ac))
* **EN-1180:** migrates to new GHA runner infra ([#752](https://github.com/doctolib/yak/issues/752)) ([4f2ac6e](https://github.com/doctolib/yak/commit/4f2ac6ee310de1c3849c124170490287442aace0))
* **EN-142:** create & delete aurora clone ([#558](https://github.com/doctolib/yak/issues/558)) ([d1e9761](https://github.com/doctolib/yak/commit/d1e9761bbbeb66208bb98d81333ae42efc60b31c))
* **EN-1501:** Declares cicd-staging Vault ns + cleanup ([#846](https://github.com/doctolib/yak/issues/846)) ([bf88d61](https://github.com/doctolib/yak/commit/bf88d6168c4e0c25d811c058bbd2c7c441d8ffd1))
* **EN-200:** Whitelist philips-labs modules ([#590](https://github.com/doctolib/yak/issues/590)) ([e6784e8](https://github.com/doctolib/yak/commit/e6784e8e0af56e72fb23c8c789e6edc1742a7a4f))
* **en-2219:** Add new repository url for github runner provider ([#920](https://github.com/doctolib/yak/issues/920)) ([6d52f13](https://github.com/doctolib/yak/commit/6d52f130eab90e34cc91e5d88d9a7392c3247457))
* **EN-337:** use rds client interface to mock client ([#602](https://github.com/doctolib/yak/issues/602)) ([62a1bd4](https://github.com/doctolib/yak/commit/62a1bd48b5542030386f83e5ee84b3e4b684ac00))
* **en-496:** add aurora cluster delete command ([#775](https://github.com/doctolib/yak/issues/775)) ([f98eb9c](https://github.com/doctolib/yak/commit/f98eb9c2d6e9f0b4615116fbd4f908dd8e837e4f))
* **EN-792:** gives ability to specify one DB Cluster parameter group per cluster created from snapshots ([#694](https://github.com/doctolib/yak/issues/694)) ([9af8cf4](https://github.com/doctolib/yak/commit/9af8cf4effcef1ddf9f5b67c89d918b527bef9b2))
* **EN-793:** Rework the way to specify target name for Aurora clones creation ([#678](https://github.com/doctolib/yak/issues/678)) ([21d8d34](https://github.com/doctolib/yak/commit/21d8d3424ca5500c8aef6685b19d4243bfaf5830))
* kafka replication monitoring ([#629](https://github.com/doctolib/yak/issues/629)) ([2db9d03](https://github.com/doctolib/yak/commit/2db9d0368d8b4e22dea86848015941c049a69a92))
* **PSRE-1355:** try to read TFE token from .terraform.d/credentials.tfrc.json ([#373](https://github.com/doctolib/yak/issues/373)) ([3140e25](https://github.com/doctolib/yak/commit/3140e259ef49944f555730b605df112e349c298c))
* **PSRE-2111:** add terraform provider publish workflow ([#669](https://github.com/doctolib/yak/issues/669)) ([b37374b](https://github.com/doctolib/yak/commit/b37374bcaa54b319f90ae947d4a465fdacc83b37))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-2730:** support renewal of certificates when DNS zone is hosted in route53 ([#935](https://github.com/doctolib/yak/issues/935)) ([9ba367f](https://github.com/doctolib/yak/commit/9ba367fa501d4f3cf67820af39b78bcd7cfb99e7))
* **PSRE-3127:** Reset workers in kafka ([#704](https://github.com/doctolib/yak/issues/704)) ([6380da8](https://github.com/doctolib/yak/commit/6380da889e34a7a2964b31c6c8fdf258aa6e7a01))
* **PSRE-3650:** prepare refactoring of vault secrets format ([#904](https://github.com/doctolib/yak/issues/904)) ([25d5a62](https://github.com/doctolib/yak/commit/25d5a620bb415feee991ae97014abeeca407cb58))
* **PSRE-3702:** Add yak command to create client and server JWT secrets ([#866](https://github.com/doctolib/yak/issues/866)) ([d8d24a8](https://github.com/doctolib/yak/commit/d8d24a81755b63c0f141c447253631c1e1a88d63))
* **PSRE-3708:** add secret undelete command ([#859](https://github.com/doctolib/yak/issues/859)) ([4097c47](https://github.com/doctolib/yak/commit/4097c47881fb9ff9288c9ebdf5f7f3cef892f5fe))
* **PSRE-3748:** Extend yak secret get command to retrieve secret data matching a provided key ([#869](https://github.com/doctolib/yak/issues/869)) ([6bae36c](https://github.com/doctolib/yak/commit/6bae36c020828ba9aae2feab950e2dc5073af7e1))
* **PSRE-4357:** yak certificate gandi-check to verify your PAT token ([81561f6](https://github.com/doctolib/yak/commit/81561f6643540154b0d5e438c1a19a7341860b83))
* **PSRE-4357:** yak certificate gandi-check to verify your PAT token ([#948](https://github.com/doctolib/yak/issues/948)) ([81561f6](https://github.com/doctolib/yak/commit/81561f6643540154b0d5e438c1a19a7341860b83))
* **PSRE-4388:** Add autogenerated file to allowed provider files ([#1082](https://github.com/doctolib/yak/issues/1082)) ([f6efa16](https://github.com/doctolib/yak/commit/f6efa16497dc413779ebe2e01e7bc0f55d727b52))
* **PSRE-4388:** Add check provider declaration file ([#943](https://github.com/doctolib/yak/issues/943)) ([c577d8a](https://github.com/doctolib/yak/commit/c577d8aeb8c88d9d8fe8b81820ce3ef94ac4c2bb))
* **PSRE-4392:** Add a check for required_providers to be in versions.tf ([#952](https://github.com/doctolib/yak/issues/952)) ([035a3f0](https://github.com/doctolib/yak/commit/035a3f0c06b5a59dfe55103b6504e3b00b3cb0f7))
* **PSRE-4397:** Add cloud declaration in backend.tf check ([#950](https://github.com/doctolib/yak/issues/950)) ([63edb72](https://github.com/doctolib/yak/commit/63edb72cce40b21f38494d56c1cda13d0107b543))
* **PSRE-4454:** get details of x509 certs stored in vault ([bc30c18](https://github.com/doctolib/yak/commit/bc30c18fffec42a0b562aa8bc19b31fe68f77ccf))
* **PSRE-4454:** get details of x509 certs stored in vault ([#947](https://github.com/doctolib/yak/issues/947)) ([bc30c18](https://github.com/doctolib/yak/commit/bc30c18fffec42a0b562aa8bc19b31fe68f77ccf))
* **PSRE-4513:** Build multi arch docker image ([#998](https://github.com/doctolib/yak/issues/998)) ([8c0033b](https://github.com/doctolib/yak/commit/8c0033bbc9f5195d50297c46cd36fc953e33ba1e))
* **PSRE-4513:** Fix yak sanity check ([#1017](https://github.com/doctolib/yak/issues/1017)) ([4af0818](https://github.com/doctolib/yak/commit/4af0818de1ae453de967dcc5aa9cb560b7dbd939))
* **PSRE-4513:** Still trying to fix docker build ([#1003](https://github.com/doctolib/yak/issues/1003)) ([b24171b](https://github.com/doctolib/yak/commit/b24171bd2eeeddb2363e5bf97221d7e805c177ee))
* **PSRE-4517:** Update Dockerfile ([#966](https://github.com/doctolib/yak/issues/966)) ([89c60e8](https://github.com/doctolib/yak/commit/89c60e8d332344c8dd5f3242b8522156fdfba324))
* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([#1066](https://github.com/doctolib/yak/issues/1066)) ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4702:** Extend yak argocd cli with additional commands ([#1124](https://github.com/doctolib/yak/issues/1124)) ([40f9037](https://github.com/doctolib/yak/commit/40f9037c853b77b0c9a642b6bb1c71f2bf56b8db))
* **PSS-1085:** support getting AWS session on new runners ([#832](https://github.com/doctolib/yak/issues/832)) ([045b35a](https://github.com/doctolib/yak/commit/045b35a29cd25b284ab3136e32230dc1e1dd6b79))
* specify latest version for module bump ([#315](https://github.com/doctolib/yak/issues/315)) ([3767bd6](https://github.com/doctolib/yak/commit/3767bd6f69788938b102cc9850d764cfda6d61d3))
* **SREGREEN-104/argocd:** support for SSO ([#735](https://github.com/doctolib/yak/issues/735)) ([784d2f4](https://github.com/doctolib/yak/commit/784d2f41a1f7029c76929f1165514f798681e9d1))
* **SREGREEN-111/helm:** add external helm charts whitelist check ([#788](https://github.com/doctolib/yak/issues/788)) ([2c33917](https://github.com/doctolib/yak/commit/2c3391773a422c8d530449685899c65e572d5719))
* **SREGREEN-111:** helm external chart whitelist check for charts ([#800](https://github.com/doctolib/yak/issues/800)) ([9e59880](https://github.com/doctolib/yak/commit/9e5988098e7b8737f3637f2b723c9e6f34972dc8))
* **SREGREEN-144:** add helm in docker image ([#786](https://github.com/doctolib/yak/issues/786)) ([0fe9346](https://github.com/doctolib/yak/commit/0fe93462287b520f98ff7aac742a2241c7c8fc4f))
* **SREGREEN-144:** yak helm export ([#737](https://github.com/doctolib/yak/issues/737)) ([f5485e0](https://github.com/doctolib/yak/commit/f5485e014e6db1b39d435d660d2eb1e0a0263f40))
* **SREGREEN-238/helm:** check if we reference prerelease version for internal charts ([#816](https://github.com/doctolib/yak/issues/816)) ([51de454](https://github.com/doctolib/yak/commit/51de4542c6533193af581c449bffece3ac6784b5))
* **SREGREEN-260:** include non-namespaced orphaned resources by default ([#850](https://github.com/doctolib/yak/issues/850)) ([1a296ac](https://github.com/doctolib/yak/commit/1a296ac7477bed5baae8fd3e630e452a3810bc68))
* **SREGREEN-281:** whitelist path where module are fetch from git ([#879](https://github.com/doctolib/yak/issues/879)) ([7023c14](https://github.com/doctolib/yak/commit/7023c144938dcb1f1090893332bc9ac13ab74d13))
* **SREGREEN-338:** Added a set-version command to workspace ([#925](https://github.com/doctolib/yak/issues/925)) ([a9c3873](https://github.com/doctolib/yak/commit/a9c38734e9e759d138a637d75e7b21a51ddc9199))
* **SREGREEN-435:** validate json in secrets ([#940](https://github.com/doctolib/yak/issues/940)) ([2b2cc92](https://github.com/doctolib/yak/commit/2b2cc92067bfc2bc124e741a81a50dfa11684e5f))
* **SREGREEN-443:** include CRDs in the helm templated output ([#922](https://github.com/doctolib/yak/issues/922)) ([9cb3c33](https://github.com/doctolib/yak/commit/9cb3c332d22a94ac67c8789dfa0b394180e40139))
* **SREGREEN-477:** add destroy command ([#930](https://github.com/doctolib/yak/issues/930)) ([c49f2c0](https://github.com/doctolib/yak/commit/c49f2c083c10c72798743097f7fb38adaf16a342))
* **SREGREEN-478:** add clean-vault command ([#951](https://github.com/doctolib/yak/issues/951)) ([b112162](https://github.com/doctolib/yak/commit/b1121629c4fcb5560ad756e94e41a13eb46f2e14))
* **SREGREEN-494:** add create-commit functionality ([#939](https://github.com/doctolib/yak/issues/939)) ([bdfc42b](https://github.com/doctolib/yak/commit/bdfc42ba33fb6f27c4497fa1fb60bf922420f5ea))
* **SREGREEN-543:** add tfeJwtSubjects in logical secrets files ([#1108](https://github.com/doctolib/yak/issues/1108)) ([1e16f7f](https://github.com/doctolib/yak/commit/1e16f7f5be40a95e2c83fe52de7c9999c5c3a89d))
* **SREGREEN-566:** add support for vaultParentNamespace in secret config ([#1053](https://github.com/doctolib/yak/issues/1053)) ([f359ed5](https://github.com/doctolib/yak/commit/f359ed5105e87ca42bb78ffb7978e271b19ef37f))
* **SREGREEN-595:** Added parallelism to couchbase logs collection ([#1064](https://github.com/doctolib/yak/issues/1064)) ([ce4179d](https://github.com/doctolib/yak/commit/ce4179ddfe5a2c39f993fa3aaee95f5d3ca05797))
* **SREGREEN-675:** add support for custom branch name in jira create-branch ([#1118](https://github.com/doctolib/yak/issues/1118)) ([50c1284](https://github.com/doctolib/yak/commit/50c128466788f5418ef953d0333da4af6aca45f1))
* **SREGREEN-675:** jira create-commit to support deriving ID from branch name ([#1119](https://github.com/doctolib/yak/issues/1119)) ([b0032b0](https://github.com/doctolib/yak/commit/b0032b04f9a035a45d289156b0065198fbc556bc))
* **SREGREEN-679:** Add helper to get env var, add YAK_JIRA_PROJECT_KEY ([#1122](https://github.com/doctolib/yak/issues/1122)) ([cf5887c](https://github.com/doctolib/yak/commit/cf5887c31dc05b0673c56bb6b5d458c56a4bb4a0))
* **SREGREEN-69:** display conditions in argocd status ([#896](https://github.com/doctolib/yak/issues/896)) ([acd33ad](https://github.com/doctolib/yak/commit/acd33ad98ebf22ef69c4227fb3b1cdacbfd4ab11))
* **teleport:** add reviewers override option ([#625](https://github.com/doctolib/yak/issues/625)) ([ab38854](https://github.com/doctolib/yak/commit/ab3885452fe82e4b4544bbcc88fba9f0800b6b11))
* **terraform report:** add terraform module report ([#350](https://github.com/doctolib/yak/issues/350)) ([a2692fc](https://github.com/doctolib/yak/commit/a2692fc435244da93d760ab4d71b45a548aa97d2))
* **tfe:** add command to discard old TFE runs - WIP ([#217](https://github.com/doctolib/yak/issues/217)) ([9c0f701](https://github.com/doctolib/yak/commit/9c0f70151a43d56d27d745b15c770384a9483d41))
* **TT-20773:** Extend yak aws aurora psql to allow for running queries against a specific rds instance ([#684](https://github.com/doctolib/yak/issues/684)) ([3d4c7c8](https://github.com/doctolib/yak/commit/3d4c7c85f08dee813058787e78734ddc9b99c2d2))


### Bug Fixes

* add container name in exec pod command ([#448](https://github.com/doctolib/yak/issues/448)) ([372b1c0](https://github.com/doctolib/yak/commit/372b1c00f9945e50d8a83b46526be520218fff6e))
* bump actions/checkout to v4 ([#394](https://github.com/doctolib/yak/issues/394)) ([099980c](https://github.com/doctolib/yak/commit/099980cf00f76938f95fda3cdeb941ad137183d0))
* **certificates:** Adapt after changes in Gandi API ([#511](https://github.com/doctolib/yak/issues/511)) ([eac085b](https://github.com/doctolib/yak/commit/eac085b95ecb2fb08e458f59878b10ca70c37ed2))
* cleaning up of ignored orphaned resources from orphaned resources was messy ([#487](https://github.com/doctolib/yak/issues/487)) ([cba3a37](https://github.com/doctolib/yak/commit/cba3a37b61ba2154454479b72fe0619e90470a16))
* **deps:** update aws-sdk-go-v2 monorepo ([#1088](https://github.com/doctolib/yak/issues/1088)) ([43366ff](https://github.com/doctolib/yak/commit/43366ff0a71288817d93c0d5936a5a236b8364fe))
* **deps:** update aws-sdk-go-v2 monorepo ([#1090](https://github.com/doctolib/yak/issues/1090)) ([3ced281](https://github.com/doctolib/yak/commit/3ced281d0dccb29a53b2a872bc4a421956607fac))
* **deps:** update aws-sdk-go-v2 monorepo ([#1099](https://github.com/doctolib/yak/issues/1099)) ([b5aa3d2](https://github.com/doctolib/yak/commit/b5aa3d25c5d1c00b6ac28a247c338d12ee20c7fe))
* **deps:** update aws-sdk-go-v2 monorepo ([#979](https://github.com/doctolib/yak/issues/979)) ([d5d76b8](https://github.com/doctolib/yak/commit/d5d76b8df9698528d6c57db33ea260de36373642))
* **deps:** update dependency go to v1.24.4 ([#1127](https://github.com/doctolib/yak/issues/1127)) ([a37ac8d](https://github.com/doctolib/yak/commit/a37ac8d6d3d7302d1faab3b2cb9731bb0882f2fc))
* **deps:** update github.com/hashicorp/terraform-config-inspect digest to d2d12f9 ([#971](https://github.com/doctolib/yak/issues/971)) ([0a002bf](https://github.com/doctolib/yak/commit/0a002bfe33dcf49549d8081557e17f4c1cb4cafe))
* **deps:** update github.com/hashicorp/terraform-config-inspect digest to f4c50e6 ([#1038](https://github.com/doctolib/yak/issues/1038)) ([7fd9ecb](https://github.com/doctolib/yak/commit/7fd9ecbc78976f18416f7d8d317d94d56e29c37a))
* **deps:** update kubernetes packages to v0.33.2 ([#980](https://github.com/doctolib/yak/issues/980)) ([a9e89fb](https://github.com/doctolib/yak/commit/a9e89fbd0b308e0ce40dc560d4114b876980dd98))
* **deps:** update module github.com/argoproj/argo-cd/v2 to v2.13.8 [security] ([#961](https://github.com/doctolib/yak/issues/961)) ([db0e5bc](https://github.com/doctolib/yak/commit/db0e5bca5eda8908515f8b17ebbc495efda59913))
* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.3 ([#973](https://github.com/doctolib/yak/issues/973)) ([86134f1](https://github.com/doctolib/yak/commit/86134f1919bcec3165c553d922817d7327ce0fb9))
* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.4 ([#1051](https://github.com/doctolib/yak/issues/1051)) ([e1d4b1a](https://github.com/doctolib/yak/commit/e1d4b1ab58feeeb61ea086932f4dd4077e27649a))
* **deps:** update module github.com/aws/aws-sdk-go to v1.55.7 ([#981](https://github.com/doctolib/yak/issues/981)) ([5d54f61](https://github.com/doctolib/yak/commit/5d54f616d39f1533b810364dc218811cc5f8e94f))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/ecr to v1.45.0 ([#1097](https://github.com/doctolib/yak/issues/1097)) ([e724949](https://github.com/doctolib/yak/commit/e724949c193afde9d5499dc0714a2b7e33e23b3e))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.96.0 ([#1054](https://github.com/doctolib/yak/issues/1054)) ([3f277d6](https://github.com/doctolib/yak/commit/3f277d6b8d23b3e738734e5f8b6c577be4ed6ddc))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.97.2 ([#1091](https://github.com/doctolib/yak/issues/1091)) ([2dc05b2](https://github.com/doctolib/yak/commit/2dc05b2aa42470ef1ee4f4a43e86b5a8f44b93c1))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.98.0 ([#1125](https://github.com/doctolib/yak/issues/1125)) ([628aa0f](https://github.com/doctolib/yak/commit/628aa0ff1e920c6c14dc6d1437a86ba01301c4e2))
* **deps:** update module github.com/aws/smithy-go to v1.22.3 ([#982](https://github.com/doctolib/yak/issues/982)) ([d0b39da](https://github.com/doctolib/yak/commit/d0b39da9f0879015d174b14a759521e2bc212f67))
* **deps:** update module github.com/aws/smithy-go to v1.22.4 ([#1095](https://github.com/doctolib/yak/issues/1095)) ([082fbb6](https://github.com/doctolib/yak/commit/082fbb65bbcaa83eec58ab5e334dc0fc28d1e72e))
* **deps:** update module github.com/birdayz/kaf to v0.2.13 ([#974](https://github.com/doctolib/yak/issues/974)) ([0f75527](https://github.com/doctolib/yak/commit/0f755279f4d54b4533d8dc86f3356849d2c9daab))
* **deps:** update module github.com/coreos/go-oidc/v3 to v3.14.1 ([#983](https://github.com/doctolib/yak/issues/983)) ([bdc5b1a](https://github.com/doctolib/yak/commit/bdc5b1ab6d24bf085b90465e26614362c36f0cf4))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.37.1 ([#984](https://github.com/doctolib/yak/issues/984)) ([e209e14](https://github.com/doctolib/yak/commit/e209e1446f631a96756dba0a981103ea2bc480e0))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.38.0 ([#1069](https://github.com/doctolib/yak/issues/1069)) ([aa76205](https://github.com/doctolib/yak/commit/aa762053a2d568ee9b5731e4461b09b327d07746))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.39.0 ([#1098](https://github.com/doctolib/yak/issues/1098)) ([e9fbb61](https://github.com/doctolib/yak/commit/e9fbb61b07400e3c590e776a3eb4ffe629cf14b2))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.40.0 ([#1121](https://github.com/doctolib/yak/issues/1121)) ([9935480](https://github.com/doctolib/yak/commit/9935480492746b652d05839aaaa1291f6229c296))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.41.0 ([#1126](https://github.com/doctolib/yak/issues/1126)) ([1a64067](https://github.com/doctolib/yak/commit/1a6406745e8492894d2264a3906a9eaf4bd15317))
* **deps:** update module github.com/fatih/color to v1.18.0 ([#985](https://github.com/doctolib/yak/issues/985)) ([931bf81](https://github.com/doctolib/yak/commit/931bf8101f30923a9987e91e1a694cd39b8123b2))
* **deps:** update module github.com/go-git/go-git/v5 to v5.13.0 [security] ([#962](https://github.com/doctolib/yak/issues/962)) ([7558f5b](https://github.com/doctolib/yak/commit/7558f5bf9f1a3277396f949577b2c4fb2bf2cc9f))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.0 ([#986](https://github.com/doctolib/yak/issues/986)) ([203a42b](https://github.com/doctolib/yak/commit/203a42beddecc6f5541323b31b6b06f34c21c7f1))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.1 ([#1077](https://github.com/doctolib/yak/issues/1077)) ([8364f4c](https://github.com/doctolib/yak/commit/8364f4ccc8d96b39b6361bd73fd5661e5955bf64))
* **deps:** update module github.com/go-git/go-git/v5 to v5.16.2 ([#1089](https://github.com/doctolib/yak/issues/1089)) ([1e3e5cc](https://github.com/doctolib/yak/commit/1e3e5cc4322f63c55585b0cdcad5daeabf400f46))
* **deps:** update module github.com/golang-jwt/jwt/v4 to v4.5.2 [security] ([#960](https://github.com/doctolib/yak/issues/960)) ([09c8038](https://github.com/doctolib/yak/commit/09c8038e45de2543fc525898bb6fd2592de1ac1f))
* **deps:** update module github.com/golang-jwt/jwt/v4 to v5 ([#1044](https://github.com/doctolib/yak/issues/1044)) ([9ce844c](https://github.com/doctolib/yak/commit/9ce844c8bf99a78e1944f60ca0b4f511cdaa50e9))
* **deps:** update module github.com/google/go-github/v52 to v72 ([#1045](https://github.com/doctolib/yak/issues/1045)) ([e04cbb0](https://github.com/doctolib/yak/commit/e04cbb0cd8146bd925420ba5be7314719b82c4b0))
* **deps:** update module github.com/hashicorp/go-tfe to v1.80.0 ([#987](https://github.com/doctolib/yak/issues/987)) ([c77a99e](https://github.com/doctolib/yak/commit/c77a99eba20280d4962438114b9e4118b9f5eb4e))
* **deps:** update module github.com/hashicorp/go-tfe to v1.81.0 ([#1058](https://github.com/doctolib/yak/issues/1058)) ([38f564c](https://github.com/doctolib/yak/commit/38f564c4bafc4f3e5342af8f3d4e29a6a8330bc1))
* **deps:** update module github.com/hashicorp/go-tfe to v1.82.0 ([#1092](https://github.com/doctolib/yak/issues/1092)) ([1631036](https://github.com/doctolib/yak/commit/16310365f6139c3781d76097684dd1fa9f3315af))
* **deps:** update module github.com/hashicorp/go-tfe to v1.83.0 ([#1100](https://github.com/doctolib/yak/issues/1100)) ([ec66302](https://github.com/doctolib/yak/commit/ec6630206980bca6140f0cb2b01e0714dce3ac18))
* **deps:** update module github.com/hashicorp/go-tfe to v1.84.0 ([#1114](https://github.com/doctolib/yak/issues/1114)) ([8407de3](https://github.com/doctolib/yak/commit/8407de3921acb9a3f01ff668b54c30882e460de2))
* **deps:** update module github.com/hashicorp/hcl/v2 to v2.23.0 ([#989](https://github.com/doctolib/yak/issues/989)) ([c47a6a5](https://github.com/doctolib/yak/commit/c47a6a5c06a82c8328e23166035a4ba7e3f4a8dc))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.14.1 ([#991](https://github.com/doctolib/yak/issues/991)) ([e0af416](https://github.com/doctolib/yak/commit/e0af4167f5d5b6ab92b4e740752b961760d143a8))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.15.0 ([#1049](https://github.com/doctolib/yak/issues/1049)) ([9feb107](https://github.com/doctolib/yak/commit/9feb1071b2d539d43524d741c44f4a972220be98))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework-jsontypes to v0.2.0 ([#992](https://github.com/doctolib/yak/issues/992)) ([015549d](https://github.com/doctolib/yak/commit/015549d3f90e934a4e2d8c456187950e262b6a69))
* **deps:** update module github.com/hashicorp/vault to v1.19.3 [security] ([#963](https://github.com/doctolib/yak/issues/963)) ([c7edbdc](https://github.com/doctolib/yak/commit/c7edbdc5af963b9492ac13da72024b5ac1404074))
* **deps:** update module github.com/hashicorp/vault to v1.19.5 ([#1112](https://github.com/doctolib/yak/issues/1112)) ([fbff5fa](https://github.com/doctolib/yak/commit/fbff5faaa1390387ec25048482b18714ffe4255f))
* **deps:** update module github.com/hashicorp/vault-plugin-secrets-kv to v0.24.1 ([#993](https://github.com/doctolib/yak/issues/993)) ([e396d19](https://github.com/doctolib/yak/commit/e396d19fdab3dcd00004d9924e21ebcff6119c9e))
* **deps:** update module github.com/hashicorp/vault/api to v1.16.0 ([#994](https://github.com/doctolib/yak/issues/994)) ([13772d9](https://github.com/doctolib/yak/commit/13772d9855d9e6f6694f5d75734514f552b7c91a))
* **deps:** update module github.com/hashicorp/vault/api to v1.20.0 ([#1078](https://github.com/doctolib/yak/issues/1078)) ([ce75d2f](https://github.com/doctolib/yak/commit/ce75d2f18e5c945fc13e2f12f19e3af11b977e09))
* **deps:** update module github.com/hashicorp/vault/sdk to v0.18.0 ([#1115](https://github.com/doctolib/yak/issues/1115)) ([3b93fd4](https://github.com/doctolib/yak/commit/3b93fd4363990c546625f1c5ec7f124efa5cd72f))
* **deps:** update module github.com/ibm/sarama to v1.45.1 ([#995](https://github.com/doctolib/yak/issues/995)) ([daf9404](https://github.com/doctolib/yak/commit/daf9404efc58112ca225b8dc288208409cb6144e))
* **deps:** update module github.com/ibm/sarama to v1.45.2 ([#1068](https://github.com/doctolib/yak/issues/1068)) ([1faef9b](https://github.com/doctolib/yak/commit/1faef9bfad695a4e687322bc01b15b50804f6a0b))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([#996](https://github.com/doctolib/yak/issues/996)) ([b556921](https://github.com/doctolib/yak/commit/b55692101288c4ce19c191d589f4da018dd04f37))
* **deps:** update module github.com/prometheus/common to v0.63.0 ([#997](https://github.com/doctolib/yak/issues/997)) ([5e22148](https://github.com/doctolib/yak/commit/5e22148fb8e8c42e811fd6ee5f7b4c9f102914b8))
* **deps:** update module github.com/prometheus/common to v0.64.0 ([#1041](https://github.com/doctolib/yak/issues/1041)) ([8d69e14](https://github.com/doctolib/yak/commit/8d69e14225313a7fd38217a682705be7c76d680a))
* **deps:** update module github.com/prometheus/common to v0.65.0 ([#1117](https://github.com/doctolib/yak/issues/1117)) ([13add5f](https://github.com/doctolib/yak/commit/13add5fcbb2cfc97ecf78283241f2661df8c3b87))
* **deps:** update module github.com/schollz/progressbar/v3 to v3.18.0 ([#1006](https://github.com/doctolib/yak/issues/1006)) ([2b60e89](https://github.com/doctolib/yak/commit/2b60e8905f198779730c2edc6fb286a9685b33ca))
* **deps:** update module github.com/spf13/cobra to v1.9.1 ([#1015](https://github.com/doctolib/yak/issues/1015)) ([96490ee](https://github.com/doctolib/yak/commit/96490ee9769c591fb813a0325815bdeb41546cf3))
* **deps:** update module github.com/spf13/viper to v1.20.1 ([#1025](https://github.com/doctolib/yak/issues/1025)) ([4594dba](https://github.com/doctolib/yak/commit/4594dba7e13c0b43f2b3bb32a417edb7d754a850))
* **deps:** update module github.com/zalando/go-keyring to v0.2.6 ([#975](https://github.com/doctolib/yak/issues/975)) ([5ef0beb](https://github.com/doctolib/yak/commit/5ef0beb1e2ff092a1d55812efe9535568392bacf))
* **deps:** update module github.com/zclconf/go-cty to v1.16.2 ([#1027](https://github.com/doctolib/yak/issues/1027)) ([0aa93db](https://github.com/doctolib/yak/commit/0aa93db625c987c794ffcc9b2fdadd535a304515))
* **deps:** update module github.com/zclconf/go-cty to v1.16.3 ([#1048](https://github.com/doctolib/yak/issues/1048)) ([639e906](https://github.com/doctolib/yak/commit/639e90656b86a9c03a85e04dd10ce66daffbe3ac))
* **deps:** update module golang.org/x/crypto to v0.35.0 [security] ([#964](https://github.com/doctolib/yak/issues/964)) ([5719d16](https://github.com/doctolib/yak/commit/5719d1651bb0ab9e4cc5d1a0e396647340c91e44))
* **deps:** update module golang.org/x/crypto to v0.38.0 ([#1028](https://github.com/doctolib/yak/issues/1028)) ([7399ae2](https://github.com/doctolib/yak/commit/7399ae2b60992e93ed2c7a0bd070bdbdc4da87f0))
* **deps:** update module golang.org/x/crypto to v0.39.0 ([#1084](https://github.com/doctolib/yak/issues/1084)) ([ef0937f](https://github.com/doctolib/yak/commit/ef0937fe8536cc3bd0333fb9960f80958039da28))
* **deps:** update module golang.org/x/mod to v0.24.0 ([#1029](https://github.com/doctolib/yak/issues/1029)) ([f46e796](https://github.com/doctolib/yak/commit/f46e79660786edac352397435ec682d429dc8ead))
* **deps:** update module golang.org/x/oauth2 to v0.30.0 ([#1030](https://github.com/doctolib/yak/issues/1030)) ([17e98c1](https://github.com/doctolib/yak/commit/17e98c1507440171614091a9dfd486611714c4f0))
* **deps:** update module golang.org/x/sync to v0.14.0 ([#1031](https://github.com/doctolib/yak/issues/1031)) ([cc3a744](https://github.com/doctolib/yak/commit/cc3a744cf2b0de119154e9640c42304c7b2e5c2a))
* **deps:** update module golang.org/x/sync to v0.15.0 ([#1086](https://github.com/doctolib/yak/issues/1086)) ([db186a6](https://github.com/doctolib/yak/commit/db186a673da677cde92c832b49a97c35b7f4396f))
* **deps:** update module golang.org/x/term to v0.32.0 ([#1033](https://github.com/doctolib/yak/issues/1033)) ([7ee03fc](https://github.com/doctolib/yak/commit/7ee03fcede537968d4d2fe8a7611ff1a2a2c938c))
* **deps:** update module sigs.k8s.io/yaml to v1.5.0 ([#1132](https://github.com/doctolib/yak/issues/1132)) ([fa0d991](https://github.com/doctolib/yak/commit/fa0d9913e90f9c07890d7b2f56a2728d5b737f56))
* **en-1244:** add workflow permissions for promote ci version workflow ([#783](https://github.com/doctolib/yak/issues/783)) ([53478d2](https://github.com/doctolib/yak/commit/53478d286b903d2f7b16cd8557a9047e4e1e0e82))
* **en-1244:** remove space ([#779](https://github.com/doctolib/yak/issues/779)) ([87956fc](https://github.com/doctolib/yak/commit/87956fc2dc57c4cf249c483faa06ce2109fc21a3))
* **EN-142:** Ensures Aurora clone uses same SecurityGroups as source ([#626](https://github.com/doctolib/yak/issues/626)) ([0140ce8](https://github.com/doctolib/yak/commit/0140ce854aaa3355fda000bb8053b36e1a00fd54))
* **EN-2538:** fixes ordering of aurora clone source/target args ([#941](https://github.com/doctolib/yak/issues/941)) ([4be2c16](https://github.com/doctolib/yak/commit/4be2c16196e11b46b2f9e479cbf47f8b939fdd19))
* in-cluster kube config ([#501](https://github.com/doctolib/yak/issues/501)) ([0ba5c35](https://github.com/doctolib/yak/commit/0ba5c35614c09c8f1265774142ec0a03a6957e4b))
* issues found by linter ([#361](https://github.com/doctolib/yak/issues/361)) ([ad29686](https://github.com/doctolib/yak/commit/ad296863b1c7b6ade6d763b65d8244d68f12d7a8))
* **kube secret check:** unexpected error after common namespace split ([#513](https://github.com/doctolib/yak/issues/513)) ([8e105bc](https://github.com/doctolib/yak/commit/8e105bc3bea6561bc5d714481be9ecffdb14d5dc))
* **provider check:** allow terraform_data resource ([#577](https://github.com/doctolib/yak/issues/577)) ([4c53f26](https://github.com/doctolib/yak/commit/4c53f262bcc62947fbe9a9f4d748d211f9bff63a))
* **PSRE-1569:** fix check command after split ([#412](https://github.com/doctolib/yak/issues/412)) ([3577c3d](https://github.com/doctolib/yak/commit/3577c3d2e876dcfa5c47f2b462b46234052426e2))
* **PSRE-2088:** add ruby to support ruby helm post-render ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-2088:** add ruby to support ruby helm post-render ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-2088:** add ruby to support ruby helm post-render ([#931](https://github.com/doctolib/yak/issues/931)) ([fb92c92](https://github.com/doctolib/yak/commit/fb92c9249fc9b1c3217bb7628ccbd068a5afcba1))
* **PSRE-3054:** use a dedicated clone for each operation ([#695](https://github.com/doctolib/yak/issues/695)) ([a6d17e3](https://github.com/doctolib/yak/commit/a6d17e38a342dfd06a8a4de63b15fd950a0763db))
* **PSRE-3748:** Fix returned secret when getting secret data with keys matching provided key ([#871](https://github.com/doctolib/yak/issues/871)) ([354c477](https://github.com/doctolib/yak/commit/354c477fb1d76b2f049af2b843e5a70e2bceacd2))
* **PSRE-3784:** Yak secret jwt server not updating CI secret ([#878](https://github.com/doctolib/yak/issues/878)) ([1adeab5](https://github.com/doctolib/yak/commit/1adeab508084f849577cf61a91435598d5955a8a))
* **PSRE-4018:** Update JWT token creation to use snake case service name in key for JWT token services ([#916](https://github.com/doctolib/yak/issues/916)) ([56ea27f](https://github.com/doctolib/yak/commit/56ea27f802ee777acb107d164f828ded31f69ed8))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4357:** remove zone from fqdn for cloudflare records ([#949](https://github.com/doctolib/yak/issues/949)) ([7c47271](https://github.com/doctolib/yak/commit/7c4727179163539638b9899590f26e9b4bd29700))
* **PSRE-4388:** Fix error count and tests ([#1034](https://github.com/doctolib/yak/issues/1034)) ([a1d6a79](https://github.com/doctolib/yak/commit/a1d6a7997415e7572903c9a959d6c0e056b7214c))
* **PSRE-4513:** Fix build of ARM image ([#1001](https://github.com/doctolib/yak/issues/1001)) ([74bc8ed](https://github.com/doctolib/yak/commit/74bc8ed956f9789f08fad69f956b536a77286ac5))
* **PSRE-4517:** fix and refactor some provider and declaration checks ([#955](https://github.com/doctolib/yak/issues/955)) ([968ebc0](https://github.com/doctolib/yak/commit/968ebc0d4c8d17fb8f4329fa221d6ae708b7b683))
* **PSRE-4517:** Update Dockerfile ([#968](https://github.com/doctolib/yak/issues/968)) ([fe8502f](https://github.com/doctolib/yak/commit/fe8502ff7fb9449001d754167161b04f6865abda))
* **PSRE-4702:** Fix for dashboard command when specifying app ([#1133](https://github.com/doctolib/yak/issues/1133)) ([0c34044](https://github.com/doctolib/yak/commit/0c3404423b14251f7f70a6f55c2637bec5001359))
* **PSRE-4702:** Refactor  argocd diff command ([#1129](https://github.com/doctolib/yak/issues/1129)) ([f430d6f](https://github.com/doctolib/yak/commit/f430d6f08cb5f743232a97488dec38f719d74345))
* **PSS-914:** Yak argocd commands failed when KUBECONFIG var is too long ([#721](https://github.com/doctolib/yak/issues/721)) ([055e79b](https://github.com/doctolib/yak/commit/055e79b59ccfab9fc4e66fca525bbc0db211e2e4))
* rename maintainer team ([#457](https://github.com/doctolib/yak/issues/457)) ([80ab7e4](https://github.com/doctolib/yak/commit/80ab7e4dd9b417fc2df40ae17de320bb00351572))
* report command does not work as expected on modules ([#351](https://github.com/doctolib/yak/issues/351)) ([33b7e3b](https://github.com/doctolib/yak/commit/33b7e3b808f677fff2a5a2d5cac57077fa536a4a))
* **secret:** terraform path fed by yak provider is not accessible ([#559](https://github.com/doctolib/yak/issues/559)) ([d811b5b](https://github.com/doctolib/yak/commit/d811b5b40f81d71c42c9d5124e5f46b97cf1d71d))
* **SREBLUE-001:** Fixes for jira and github commands ([#645](https://github.com/doctolib/yak/issues/645)) ([1eb2739](https://github.com/doctolib/yak/commit/1eb273929b80c4440e75b0c1da6cf2e04fabaaa6))
* **SREGREEN-111:** check for chart is incomplete ([#801](https://github.com/doctolib/yak/issues/801)) ([488aa09](https://github.com/doctolib/yak/commit/488aa09b05545d6ee2f2bbc4f5d1dd506193c412))
* **SREGREEN-111:** incomplete error message ([#805](https://github.com/doctolib/yak/issues/805)) ([b868fe3](https://github.com/doctolib/yak/commit/b868fe30b4ad8c092e70975bf2c5418e2bd626ce))
* **SREGREEN-111:** panic when there is no dependency ([#804](https://github.com/doctolib/yak/issues/804)) ([54b85f0](https://github.com/doctolib/yak/commit/54b85f03c715ac69c1aa6cfd2e4b878ed7b16174))
* **SREGREEN-144:** fix config bug and add tests ([#784](https://github.com/doctolib/yak/issues/784)) ([62a7d5e](https://github.com/doctolib/yak/commit/62a7d5eaf5789c63d382ad15cb302a7937ef52a3))
* **SREGREEN-144:** fix managed files ([#787](https://github.com/doctolib/yak/issues/787)) ([82f75b8](https://github.com/doctolib/yak/commit/82f75b83dd4f7db5bd24f0853afc08ef415662ee))
* **SREGREEN-215:** add bybasstsh flag deleted by mistake ([#778](https://github.com/doctolib/yak/issues/778)) ([f4328e3](https://github.com/doctolib/yak/commit/f4328e3fd33e73e3465c5522cc38f318db16a91c))
* **SREGREEN-221:** helm template printing logs to stdout... ([#790](https://github.com/doctolib/yak/issues/790)) ([a50bdb3](https://github.com/doctolib/yak/commit/a50bdb342ea3a14a7e43b2e2301057a1df29de66))
* **SREGREEN-228:** lags calculation ([#820](https://github.com/doctolib/yak/issues/820)) ([1206ea0](https://github.com/doctolib/yak/commit/1206ea0689a943f4b08b652b177aed8bb86db67a))
* **SREGREEN-350:** automate common semantic release flow ([#953](https://github.com/doctolib/yak/issues/953)) ([cde76a9](https://github.com/doctolib/yak/commit/cde76a9f6111c86165bce1b1e3bbb23914717e3e))
* **SREGREEN-350:** configure renovate to run `go mod tidy` ([#1050](https://github.com/doctolib/yak/issues/1050)) ([bdb1989](https://github.com/doctolib/yak/commit/bdb198905c9b6dc4f8a3c4c7c42c154ecf67920e))
* **SREGREEN-350:** use beefier runner for publish workflow ([#959](https://github.com/doctolib/yak/issues/959)) ([85236c0](https://github.com/doctolib/yak/commit/85236c0526b482a3f9bc8b2ba610eba3da4ff232))
* **SREGREEN-449:** --config flag not working ([#923](https://github.com/doctolib/yak/issues/923)) ([2506b17](https://github.com/doctolib/yak/commit/2506b17bdd5f1af482af13911e2e3ff7b0d25fb8))
* **SREGREEN-48:** update argocd cli to use "main" project ([#641](https://github.com/doctolib/yak/issues/641)) ([d0e134f](https://github.com/doctolib/yak/commit/d0e134f6ac1ce385c3554809e94ab6bf7e6af750))
* **SREGREEN-499:** create_branch: do not clean local index + hide completed tasks ([#944](https://github.com/doctolib/yak/issues/944)) ([c17ce75](https://github.com/doctolib/yak/commit/c17ce7557455e35554b8a14ead7bde611a841380))
* **SREGREEN-575:** honor the -a option in 'yak argocd status' ([#1062](https://github.com/doctolib/yak/issues/1062)) ([dda7cb0](https://github.com/doctolib/yak/commit/dda7cb0ea7c95042028e2209b7c5b0545153b6f1))
* **SREGREEN-57:** sort alphabetically output of yak argocd status ([#847](https://github.com/doctolib/yak/issues/847)) ([d8a5b06](https://github.com/doctolib/yak/commit/d8a5b06a939de10ec7fc5af07cf487006d144609))
* **SREGREEN-635:** argocd suspend to support UI suspensions ([#1072](https://github.com/doctolib/yak/issues/1072)) ([e0d2fc0](https://github.com/doctolib/yak/commit/e0d2fc0fc3201a2a0c3c0b7c18cbe0185967afd4))
* **SREGREEN-640/jwt:** config flag not being honored ([#1080](https://github.com/doctolib/yak/issues/1080)) ([57c6512](https://github.com/doctolib/yak/commit/57c6512a792b395c70e77e2e2465982f77fa4cc7))
* **SREGREEN-70/argocd:** fix rendering of argocd status ([#761](https://github.com/doctolib/yak/issues/761)) ([4eca3e0](https://github.com/doctolib/yak/commit/4eca3e063cf3d4353893bd057c9f1d25f278206d))
* use logrus lib instead of default lib ([#483](https://github.com/doctolib/yak/issues/483)) ([f3b3bbf](https://github.com/doctolib/yak/commit/f3b3bbf29bd4e1d39132d6f66c0285f1ff531c33))
* wrong team assignation with check-versions ([#349](https://github.com/doctolib/yak/issues/349)) ([6b68c9d](https://github.com/doctolib/yak/commit/6b68c9d378674680adbdcad8a68bf3175d24837b))


### Reverts

* **EN-1035:** revert change on publish.yml ([#733](https://github.com/doctolib/yak/issues/733)) ([5fbf1a3](https://github.com/doctolib/yak/commit/5fbf1a370b5f42b98ff2e11d239f448022309a8a))


### Code Refactoring

* **PSRE-3650:** support new format for vault-secrets files ([#902](https://github.com/doctolib/yak/issues/902)) ([856084a](https://github.com/doctolib/yak/commit/856084a5aef6156af4fffe7e6050375f8bf5a5eb))

## [2.21.3](https://github.com/doctolib/yak/compare/v2.21.2...v2.21.3) (2025-06-26)


### Bug Fixes

* **deps:** update dependency go to v1.24.4 ([#1127](https://github.com/doctolib/yak/issues/1127)) ([a37ac8d](https://github.com/doctolib/yak/commit/a37ac8d6d3d7302d1faab3b2cb9731bb0882f2fc))

## [2.21.2](https://github.com/doctolib/yak/compare/v2.21.1...v2.21.2) (2025-06-26)


### Bug Fixes

* **deps:** update module sigs.k8s.io/yaml to v1.5.0 ([#1132](https://github.com/doctolib/yak/issues/1132)) ([fa0d991](https://github.com/doctolib/yak/commit/fa0d9913e90f9c07890d7b2f56a2728d5b737f56))
* **PSRE-4702:** Fix for dashboard command when specifying app ([#1133](https://github.com/doctolib/yak/issues/1133)) ([0c34044](https://github.com/doctolib/yak/commit/0c3404423b14251f7f70a6f55c2637bec5001359))

## [2.20.0](https://github.com/doctolib/yak/compare/v2.19.0...v2.20.0) (2025-06-24)


### Features

* **SREGREEN-675:** add support for custom branch name in jira create-branch ([#1118](https://github.com/doctolib/yak/issues/1118)) ([50c1284](https://github.com/doctolib/yak/commit/50c128466788f5418ef953d0333da4af6aca45f1))
* **SREGREEN-675:** jira create-commit to support deriving ID from branch name ([#1119](https://github.com/doctolib/yak/issues/1119)) ([b0032b0](https://github.com/doctolib/yak/commit/b0032b04f9a035a45d289156b0065198fbc556bc))


### Bug Fixes

* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.40.0 ([#1121](https://github.com/doctolib/yak/issues/1121)) ([9935480](https://github.com/doctolib/yak/commit/9935480492746b652d05839aaaa1291f6229c296))
* **deps:** update module github.com/hashicorp/go-tfe to v1.84.0 ([#1114](https://github.com/doctolib/yak/issues/1114)) ([8407de3](https://github.com/doctolib/yak/commit/8407de3921acb9a3f01ff668b54c30882e460de2))
* **deps:** update module github.com/hashicorp/vault to v1.19.5 ([#1112](https://github.com/doctolib/yak/issues/1112)) ([fbff5fa](https://github.com/doctolib/yak/commit/fbff5faaa1390387ec25048482b18714ffe4255f))
* **deps:** update module github.com/hashicorp/vault/sdk to v0.18.0 ([#1115](https://github.com/doctolib/yak/issues/1115)) ([3b93fd4](https://github.com/doctolib/yak/commit/3b93fd4363990c546625f1c5ec7f124efa5cd72f))
* **deps:** update module github.com/prometheus/common to v0.65.0 ([#1117](https://github.com/doctolib/yak/issues/1117)) ([13add5f](https://github.com/doctolib/yak/commit/13add5fcbb2cfc97ecf78283241f2661df8c3b87))

## [2.19.0](https://github.com/doctolib/yak/compare/v2.18.7...v2.19.0) (2025-06-20)


### Features

* **SREGREEN-543:** add tfeJwtSubjects in logical secrets files ([#1108](https://github.com/doctolib/yak/issues/1108)) ([1e16f7f](https://github.com/doctolib/yak/commit/1e16f7f5be40a95e2c83fe52de7c9999c5c3a89d))


### Bug Fixes

* **deps:** update module github.com/hashicorp/vault-plugin-secrets-kv to v0.24.1 ([#993](https://github.com/doctolib/yak/issues/993)) ([e396d19](https://github.com/doctolib/yak/commit/e396d19fdab3dcd00004d9924e21ebcff6119c9e))

## [2.18.7](https://github.com/doctolib/yak/compare/v2.18.6...v2.18.7) (2025-06-20)


### Bug Fixes

* **deps:** update module github.com/hashicorp/vault to v1.19.3 [security] ([#963](https://github.com/doctolib/yak/issues/963)) ([c7edbdc](https://github.com/doctolib/yak/commit/c7edbdc5af963b9492ac13da72024b5ac1404074))

## [2.18.6](https://github.com/doctolib/yak/compare/v2.18.5...v2.18.6) (2025-06-20)


### Bug Fixes

* **deps:** update kubernetes packages to v0.33.2 ([#980](https://github.com/doctolib/yak/issues/980)) ([a9e89fb](https://github.com/doctolib/yak/commit/a9e89fbd0b308e0ce40dc560d4114b876980dd98))

## [2.18.5](https://github.com/doctolib/yak/compare/v2.18.4...v2.18.5) (2025-06-20)


### Bug Fixes

* **deps:** update module github.com/argoproj/argo-cd/v2 to v2.13.8 [security] ([#961](https://github.com/doctolib/yak/issues/961)) ([db0e5bc](https://github.com/doctolib/yak/commit/db0e5bca5eda8908515f8b17ebbc495efda59913))

## [2.18.4](https://github.com/doctolib/yak/compare/v2.18.3...v2.18.4) (2025-06-19)


### Dependency Updates

* **deps:** go-github: `MergeMethod*` consts have been split into: `PullRequestMergeMethod*` and `MergeQueueMergeMethod*`.
    - feat: Add support for pagination options in rules API methods
    ([#&#8203;3562](https://redirect.github.com/google/go-github/issues/3562))
    `GetRulesForBranch`, `GetAllRulesets`, and
    `GetAllRepositoryRulesets` now accept `opts`.

### Bug Fixes

* **deps:** update module github.com/google/go-github/v52 to v72 ([#1045](https://github.com/doctolib/yak/issues/1045)) ([e04cbb0](https://github.com/doctolib/yak/commit/e04cbb0cd8146bd925420ba5be7314719b82c4b0))

## [2.18.3](https://github.com/doctolib/yak/compare/v2.18.2...v2.18.3) (2025-06-18)


### Bug Fixes

* **deps:** update aws-sdk-go-v2 monorepo ([#1099](https://github.com/doctolib/yak/issues/1099)) ([b5aa3d2](https://github.com/doctolib/yak/commit/b5aa3d25c5d1c00b6ac28a247c338d12ee20c7fe))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/ecr to v1.45.0 ([#1097](https://github.com/doctolib/yak/issues/1097)) ([e724949](https://github.com/doctolib/yak/commit/e724949c193afde9d5499dc0714a2b7e33e23b3e))
* **deps:** update module github.com/aws/smithy-go to v1.22.4 ([#1095](https://github.com/doctolib/yak/issues/1095)) ([082fbb6](https://github.com/doctolib/yak/commit/082fbb65bbcaa83eec58ab5e334dc0fc28d1e72e))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.39.0 ([#1098](https://github.com/doctolib/yak/issues/1098)) ([e9fbb61](https://github.com/doctolib/yak/commit/e9fbb61b07400e3c590e776a3eb4ffe629cf14b2))
* **deps:** update module github.com/hashicorp/go-tfe to v1.83.0 ([#1100](https://github.com/doctolib/yak/issues/1100)) ([ec66302](https://github.com/doctolib/yak/commit/ec6630206980bca6140f0cb2b01e0714dce3ac18))

## [2.18.2](https://github.com/doctolib/yak/compare/v2.18.1...v2.18.2) (2025-06-13)


### Bug Fixes

* **deps:** update module github.com/hashicorp/go-tfe to v1.82.0 ([#1092](https://github.com/doctolib/yak/issues/1092)) ([1631036](https://github.com/doctolib/yak/commit/16310365f6139c3781d76097684dd1fa9f3315af))

## [2.18.1](https://github.com/doctolib/yak/compare/v2.18.0...v2.18.1) (2025-06-12)


### Bug Fixes

* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.97.2 ([#1091](https://github.com/doctolib/yak/issues/1091)) ([2dc05b2](https://github.com/doctolib/yak/commit/2dc05b2aa42470ef1ee4f4a43e86b5a8f44b93c1))

## [2.17.1](https://github.com/doctolib/yak/compare/v2.17.0...v2.17.1) (2025-06-05)


### Bug Fixes

* **deps:** update module github.com/go-git/go-git/v5 to v5.16.1 ([#1077](https://github.com/doctolib/yak/issues/1077)) ([8364f4c](https://github.com/doctolib/yak/commit/8364f4ccc8d96b39b6361bd73fd5661e5955bf64))
* **SREGREEN-640/jwt:** config flag not being honored ([#1080](https://github.com/doctolib/yak/issues/1080)) ([57c6512](https://github.com/doctolib/yak/commit/57c6512a792b395c70e77e2e2465982f77fa4cc7))

## [2.17.0](https://github.com/doctolib/yak/compare/v2.16.0...v2.17.0) (2025-06-04)


### Features

* **SREGREEN-595:** Added parallelism to couchbase logs collection ([#1064](https://github.com/doctolib/yak/issues/1064)) ([ce4179d](https://github.com/doctolib/yak/commit/ce4179ddfe5a2c39f993fa3aaee95f5d3ca05797))


### Bug Fixes

* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.38.0 ([#1069](https://github.com/doctolib/yak/issues/1069)) ([aa76205](https://github.com/doctolib/yak/commit/aa762053a2d568ee9b5731e4461b09b327d07746))
* **deps:** update module github.com/ibm/sarama to v1.45.2 ([#1068](https://github.com/doctolib/yak/issues/1068)) ([1faef9b](https://github.com/doctolib/yak/commit/1faef9bfad695a4e687322bc01b15b50804f6a0b))
* **SREGREEN-635:** argocd suspend to support UI suspensions ([#1072](https://github.com/doctolib/yak/issues/1072)) ([e0d2fc0](https://github.com/doctolib/yak/commit/e0d2fc0fc3201a2a0c3c0b7c18cbe0185967afd4))

## [2.16.0](https://github.com/doctolib/yak/compare/v2.15.2...v2.16.0) (2025-05-28)


### Features

* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))
* **PSRE-4654:** release please ([#1066](https://github.com/doctolib/yak/issues/1066)) ([2e23f4a](https://github.com/doctolib/yak/commit/2e23f4a5db4347745a0190993698f972fc414cea))

## [2.15.2](https://github.com/doctolib/yak/compare/v2.15.1...v2.15.2) (2025-05-27)


### Bug Fixes

* **SREGREEN-575:** honor the -a option in 'yak argocd status' ([#1062](https://github.com/doctolib/yak/issues/1062)) ([dda7cb0](https://github.com/doctolib/yak/commit/dda7cb0ea7c95042028e2209b7c5b0545153b6f1))

## [2.15.1](https://github.com/doctolib/yak/compare/v2.15.0...v2.15.1) (2025-05-27)


### Bug Fixes

* **deps:** update module github.com/golang-jwt/jwt/v4 to v5 ([#1044](https://github.com/doctolib/yak/issues/1044)) ([9ce844c](https://github.com/doctolib/yak/commit/9ce844c8bf99a78e1944f60ca0b4f511cdaa50e9))

## [2.15.0](https://github.com/doctolib/yak/compare/v2.14.7...v2.15.0) (2025-05-26)


### Features

* **SREGREEN-566:** add support for vaultParentNamespace in secret config ([#1053](https://github.com/doctolib/yak/issues/1053)) ([f359ed5](https://github.com/doctolib/yak/commit/f359ed5105e87ca42bb78ffb7978e271b19ef37f))


### Bug Fixes

* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.4 ([#1051](https://github.com/doctolib/yak/issues/1051)) ([e1d4b1a](https://github.com/doctolib/yak/commit/e1d4b1ab58feeeb61ea086932f4dd4077e27649a))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/rds to v1.96.0 ([#1054](https://github.com/doctolib/yak/issues/1054)) ([3f277d6](https://github.com/doctolib/yak/commit/3f277d6b8d23b3e738734e5f8b6c577be4ed6ddc))
* **deps:** update module github.com/hashicorp/go-tfe to v1.81.0 ([#1058](https://github.com/doctolib/yak/issues/1058)) ([38f564c](https://github.com/doctolib/yak/commit/38f564c4bafc4f3e5342af8f3d4e29a6a8330bc1))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.15.0 ([#1049](https://github.com/doctolib/yak/issues/1049)) ([9feb107](https://github.com/doctolib/yak/commit/9feb1071b2d539d43524d741c44f4a972220be98))

## [2.14.7](https://github.com/doctolib/yak/compare/v2.14.6...v2.14.7) (2025-05-19)


### Bug Fixes

* **deps:** update github.com/hashicorp/terraform-config-inspect digest to f4c50e6 ([#1038](https://github.com/doctolib/yak/issues/1038)) ([7fd9ecb](https://github.com/doctolib/yak/commit/7fd9ecbc78976f18416f7d8d317d94d56e29c37a))
* **deps:** update module github.com/hashicorp/go-tfe to v1.80.0 ([#987](https://github.com/doctolib/yak/issues/987)) ([c77a99e](https://github.com/doctolib/yak/commit/c77a99eba20280d4962438114b9e4118b9f5eb4e))
* **deps:** update module github.com/prometheus/common to v0.64.0 ([#1041](https://github.com/doctolib/yak/issues/1041)) ([8d69e14](https://github.com/doctolib/yak/commit/8d69e14225313a7fd38217a682705be7c76d680a))
* **deps:** update module github.com/zclconf/go-cty to v1.16.3 ([#1048](https://github.com/doctolib/yak/issues/1048)) ([639e906](https://github.com/doctolib/yak/commit/639e90656b86a9c03a85e04dd10ce66daffbe3ac))
* **SREGREEN-350:** configure renovate to run `go mod tidy` ([#1050](https://github.com/doctolib/yak/issues/1050)) ([bdb1989](https://github.com/doctolib/yak/commit/bdb198905c9b6dc4f8a3c4c7c42c154ecf67920e))

## [2.14.6](https://github.com/doctolib/yak/compare/v2.14.5...v2.14.6) (2025-05-15)


### Bug Fixes

* **PSRE-4388:** Fix error count and tests ([#1034](https://github.com/doctolib/yak/issues/1034)) ([a1d6a79](https://github.com/doctolib/yak/commit/a1d6a7997415e7572903c9a959d6c0e056b7214c))

## [2.14.5](https://github.com/doctolib/yak/compare/v2.14.4...v2.14.5) (2025-05-15)


### Bug Fixes

* **deps:** update module github.com/hashicorp/terraform-plugin-framework to v1.14.1 ([#991](https://github.com/doctolib/yak/issues/991)) ([e0af416](https://github.com/doctolib/yak/commit/e0af4167f5d5b6ab92b4e740752b961760d143a8))
* **deps:** update module github.com/schollz/progressbar/v3 to v3.18.0 ([#1006](https://github.com/doctolib/yak/issues/1006)) ([2b60e89](https://github.com/doctolib/yak/commit/2b60e8905f198779730c2edc6fb286a9685b33ca))
* **deps:** update module github.com/spf13/cobra to v1.9.1 ([#1015](https://github.com/doctolib/yak/issues/1015)) ([96490ee](https://github.com/doctolib/yak/commit/96490ee9769c591fb813a0325815bdeb41546cf3))
* **deps:** update module github.com/spf13/viper to v1.20.1 ([#1025](https://github.com/doctolib/yak/issues/1025)) ([4594dba](https://github.com/doctolib/yak/commit/4594dba7e13c0b43f2b3bb32a417edb7d754a850))
* **deps:** update module github.com/zclconf/go-cty to v1.16.2 ([#1027](https://github.com/doctolib/yak/issues/1027)) ([0aa93db](https://github.com/doctolib/yak/commit/0aa93db625c987c794ffcc9b2fdadd535a304515))
* **deps:** update module golang.org/x/crypto to v0.38.0 ([#1028](https://github.com/doctolib/yak/issues/1028)) ([7399ae2](https://github.com/doctolib/yak/commit/7399ae2b60992e93ed2c7a0bd070bdbdc4da87f0))
* **deps:** update module golang.org/x/mod to v0.24.0 ([#1029](https://github.com/doctolib/yak/issues/1029)) ([f46e796](https://github.com/doctolib/yak/commit/f46e79660786edac352397435ec682d429dc8ead))
* **deps:** update module golang.org/x/oauth2 to v0.30.0 ([#1030](https://github.com/doctolib/yak/issues/1030)) ([17e98c1](https://github.com/doctolib/yak/commit/17e98c1507440171614091a9dfd486611714c4f0))
* **deps:** update module golang.org/x/sync to v0.14.0 ([#1031](https://github.com/doctolib/yak/issues/1031)) ([cc3a744](https://github.com/doctolib/yak/commit/cc3a744cf2b0de119154e9640c42304c7b2e5c2a))
* **deps:** update module golang.org/x/term to v0.32.0 ([#1033](https://github.com/doctolib/yak/issues/1033)) ([7ee03fc](https://github.com/doctolib/yak/commit/7ee03fcede537968d4d2fe8a7611ff1a2a2c938c))

## [2.14.4](https://github.com/doctolib/yak/compare/v2.14.3...v2.14.4) (2025-05-14)


### Bug Fixes

* **deps:** update module github.com/hashicorp/hcl/v2 to v2.23.0 ([#989](https://github.com/doctolib/yak/issues/989)) ([c47a6a5](https://github.com/doctolib/yak/commit/c47a6a5c06a82c8328e23166035a4ba7e3f4a8dc))
* **deps:** update module github.com/hashicorp/terraform-plugin-framework-jsontypes to v0.2.0 ([#992](https://github.com/doctolib/yak/issues/992)) ([015549d](https://github.com/doctolib/yak/commit/015549d3f90e934a4e2d8c456187950e262b6a69))
* **deps:** update module github.com/hashicorp/vault/api to v1.16.0 ([#994](https://github.com/doctolib/yak/issues/994)) ([13772d9](https://github.com/doctolib/yak/commit/13772d9855d9e6f6694f5d75734514f552b7c91a))
* **deps:** update module github.com/ibm/sarama to v1.45.1 ([#995](https://github.com/doctolib/yak/issues/995)) ([daf9404](https://github.com/doctolib/yak/commit/daf9404efc58112ca225b8dc288208409cb6144e))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([#996](https://github.com/doctolib/yak/issues/996)) ([b556921](https://github.com/doctolib/yak/commit/b55692101288c4ce19c191d589f4da018dd04f37))
* **deps:** update module github.com/prometheus/common to v0.63.0 ([#997](https://github.com/doctolib/yak/issues/997)) ([5e22148](https://github.com/doctolib/yak/commit/5e22148fb8e8c42e811fd6ee5f7b4c9f102914b8))

## [2.14.3](https://github.com/doctolib/yak/compare/v2.14.2...v2.14.3) (2025-05-12)


### Bug Fixes

* **PSRE-4513:** Fix yak sanity check ([#1017](https://github.com/doctolib/yak/issues/1017)) ([4af0818](https://github.com/doctolib/yak/commit/4af0818de1ae453de967dcc5aa9cb560b7dbd939))

## [2.14.2](https://github.com/doctolib/yak/compare/v2.14.1...v2.14.2) (2025-05-12)


### Bug Fixes

* **deps:** update module github.com/go-git/go-git/v5 to v5.16.0 ([#986](https://github.com/doctolib/yak/issues/986)) ([203a42b](https://github.com/doctolib/yak/commit/203a42beddecc6f5541323b31b6b06f34c21c7f1))

## [2.14.1](https://github.com/doctolib/yak/compare/v2.14.0...v2.14.1) (2025-05-09)


### Bug Fixes

* **deps:** update module github.com/aws/aws-msk-iam-sasl-signer-go to v1.0.3 ([#973](https://github.com/doctolib/yak/issues/973)) ([86134f1](https://github.com/doctolib/yak/commit/86134f1919bcec3165c553d922817d7327ce0fb9))
* **deps:** update module github.com/aws/aws-sdk-go to v1.55.7 ([#981](https://github.com/doctolib/yak/issues/981)) ([5d54f61](https://github.com/doctolib/yak/commit/5d54f616d39f1533b810364dc218811cc5f8e94f))
* **deps:** update module github.com/aws/smithy-go to v1.22.3 ([#982](https://github.com/doctolib/yak/issues/982)) ([d0b39da](https://github.com/doctolib/yak/commit/d0b39da9f0879015d174b14a759521e2bc212f67))
* **deps:** update module github.com/coreos/go-oidc/v3 to v3.14.1 ([#983](https://github.com/doctolib/yak/issues/983)) ([bdc5b1a](https://github.com/doctolib/yak/commit/bdc5b1ab6d24bf085b90465e26614362c36f0cf4))
* **deps:** update module github.com/datadog/datadog-api-client-go/v2 to v2.37.1 ([#984](https://github.com/doctolib/yak/issues/984)) ([e209e14](https://github.com/doctolib/yak/commit/e209e1446f631a96756dba0a981103ea2bc480e0))
* **deps:** update module github.com/fatih/color to v1.18.0 ([#985](https://github.com/doctolib/yak/issues/985)) ([931bf81](https://github.com/doctolib/yak/commit/931bf8101f30923a9987e91e1a694cd39b8123b2))

## [2.14.0](https://github.com/doctolib/yak/compare/v2.13.1...v2.14.0) (2025-05-07)


### Features

* **PSRE-4513:** Still trying to fix docker build ([#1003](https://github.com/doctolib/yak/issues/1003)) ([b24171b](https://github.com/doctolib/yak/commit/b24171bd2eeeddb2363e5bf97221d7e805c177ee))

## [2.13.1](https://github.com/doctolib/yak/compare/v2.13.0...v2.13.1) (2025-05-07)


### Bug Fixes

* **PSRE-4513:** Fix build of ARM image ([#1001](https://github.com/doctolib/yak/issues/1001)) ([74bc8ed](https://github.com/doctolib/yak/commit/74bc8ed956f9789f08fad69f956b536a77286ac5))

## [2.13.0](https://github.com/doctolib/yak/compare/v2.12.1...v2.13.0) (2025-05-07)


### Features

* **PSRE-4513:** Build multi arch docker image ([#998](https://github.com/doctolib/yak/issues/998)) ([8c0033b](https://github.com/doctolib/yak/commit/8c0033bbc9f5195d50297c46cd36fc953e33ba1e))


### Bug Fixes

* **deps:** update aws-sdk-go-v2 monorepo ([#979](https://github.com/doctolib/yak/issues/979)) ([d5d76b8](https://github.com/doctolib/yak/commit/d5d76b8df9698528d6c57db33ea260de36373642))

## [2.12.1](https://github.com/doctolib/yak/compare/v2.12.0...v2.12.1) (2025-05-06)


### Bug Fixes

* **deps:** update github.com/hashicorp/terraform-config-inspect digest to d2d12f9 ([#971](https://github.com/doctolib/yak/issues/971)) ([0a002bf](https://github.com/doctolib/yak/commit/0a002bfe33dcf49549d8081557e17f4c1cb4cafe))
* **deps:** update module github.com/birdayz/kaf to v0.2.13 ([#974](https://github.com/doctolib/yak/issues/974)) ([0f75527](https://github.com/doctolib/yak/commit/0f755279f4d54b4533d8dc86f3356849d2c9daab))
* **deps:** update module github.com/zalando/go-keyring to v0.2.6 ([#975](https://github.com/doctolib/yak/issues/975)) ([5ef0beb](https://github.com/doctolib/yak/commit/5ef0beb1e2ff092a1d55812efe9535568392bacf))

## [2.12.0](https://github.com/doctolib/yak/compare/v2.11.2...v2.12.0) (2025-05-05)


### Features

* **PSRE-4517:** Update Dockerfile ([#966](https://github.com/doctolib/yak/issues/966)) ([89c60e8](https://github.com/doctolib/yak/commit/89c60e8d332344c8dd5f3242b8522156fdfba324))


### Bug Fixes

* **deps:** update module golang.org/x/crypto to v0.35.0 [security] ([#964](https://github.com/doctolib/yak/issues/964)) ([5719d16](https://github.com/doctolib/yak/commit/5719d1651bb0ab9e4cc5d1a0e396647340c91e44))
* **PSRE-4517:** Update Dockerfile ([#968](https://github.com/doctolib/yak/issues/968)) ([fe8502f](https://github.com/doctolib/yak/commit/fe8502ff7fb9449001d754167161b04f6865abda))

## [2.11.2](https://github.com/doctolib/yak/compare/v2.11.1...v2.11.2) (2025-05-05)


### Bug Fixes

* **deps:** update module github.com/go-git/go-git/v5 to v5.13.0 [security] ([#962](https://github.com/doctolib/yak/issues/962)) ([7558f5b](https://github.com/doctolib/yak/commit/7558f5bf9f1a3277396f949577b2c4fb2bf2cc9f))
* **deps:** update module github.com/golang-jwt/jwt/v4 to v4.5.2 [security] ([#960](https://github.com/doctolib/yak/issues/960)) ([09c8038](https://github.com/doctolib/yak/commit/09c8038e45de2543fc525898bb6fd2592de1ac1f))
* **SREGREEN-350:** use beefier runner for publish workflow ([#959](https://github.com/doctolib/yak/issues/959)) ([85236c0](https://github.com/doctolib/yak/commit/85236c0526b482a3f9bc8b2ba610eba3da4ff232))

## [2.11.1](https://github.com/doctolib/yak/compare/v2.11.0...v2.11.1) (2025-05-05)


### Bug Fixes

* **PSRE-4517:** fix and refactor some provider and declaration checks ([#955](https://github.com/doctolib/yak/issues/955)) ([968ebc0](https://github.com/doctolib/yak/commit/968ebc0d4c8d17fb8f4329fa221d6ae708b7b683))
* **SREGREEN-350:** automate common semantic release flow ([#953](https://github.com/doctolib/yak/issues/953)) ([cde76a9](https://github.com/doctolib/yak/commit/cde76a9f6111c86165bce1b1e3bbb23914717e3e))
