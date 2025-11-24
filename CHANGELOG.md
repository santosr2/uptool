# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes
- Ci, go.mod and golangci by [@santosr2](https://github.com/santosr2) ([a242141](https://github.com/santosr2/uptool/commit/a24214179371307d7bcf0c55ae49ed3d10a270d7))
- Remove docs symlinks since it breaks action cache by [@santosr2](https://github.com/santosr2) ([109da98](https://github.com/santosr2/uptool/commit/109da985871eaffba17f0195983161c5237e0373))
- Linters and CI by [@santosr2](https://github.com/santosr2) ([9cb444f](https://github.com/santosr2/uptool/commit/9cb444ffd30ab9a68f52db50aec9712818b41134))

### Container
- Improve based on linter by [@santosr2](https://github.com/santosr2) ([97345cf](https://github.com/santosr2/uptool/commit/97345cf5b3e0e5495e13bd9a97292a92ec730177))
- Use go1.25 base image in builder stage by [@santosr2](https://github.com/santosr2) ([1112efb](https://github.com/santosr2/uptool/commit/1112efb7419af8b4e5d1217dfc872962e2e3153f))

### Continuous Integration

- **changelog**: Ensure actors are tagged by [@santosr2](https://github.com/santosr2)([94841e0](https://github.com/santosr2/uptool/commit/94841e0ca238330c9a453da04fce84a8cc856471))

- **docker**: Fix scan command by [@santosr2](https://github.com/santosr2)([0931d34](https://github.com/santosr2/uptool/commit/0931d34503c04c9ab992dc82d264eb648a84267c))

- **docker**: Fix image build by [@santosr2](https://github.com/santosr2)([ee1d558](https://github.com/santosr2/uptool/commit/ee1d5580cef49eade347f4b5e562180436c74607))

- **pre-release**: Remove unnecessary sed command by [@santosr2](https://github.com/santosr2)([7b2a8f2](https://github.com/santosr2/uptool/commit/7b2a8f2ee98fd31abfdb460afe9f3310cde3e959))

- **pre-release**: Fix checkout strategy by [@santosr2](https://github.com/santosr2)([b5bc171](https://github.com/santosr2/uptool/commit/b5bc17174852cb1f3e6b10559da3b13156fcdd54))

- **release**: Remove C dependency by [@santosr2](https://github.com/santosr2)([46c7b0b](https://github.com/santosr2/uptool/commit/46c7b0b6dce9e30a6494167cecb5183e654ffab8))

- **signed-commits**: Ensure it fails when no-signed commits are found by [@santosr2](https://github.com/santosr2)([7d03c45](https://github.com/santosr2/uptool/commit/7d03c45841f7fea59b23c06b08f9c2eeeb9bd194))
- Add ci pipeline by [@santosr2](https://github.com/santosr2) ([627232e](https://github.com/santosr2/uptool/commit/627232e27ecde493434aa1ad3c46f1c9ad628b86))
- Add doc deploy by [@santosr2](https://github.com/santosr2) ([e113950](https://github.com/santosr2/uptool/commit/e1139507d378cda3a2adcc0256111cf461e29c11))
- Fix doc workflow by [@santosr2](https://github.com/santosr2) ([e2cb1ee](https://github.com/santosr2/uptool/commit/e2cb1ee4a2351c779f5bb6c9a82cac28951f87e3))
- Fix golangci-lint by [@santosr2](https://github.com/santosr2) ([9c4be34](https://github.com/santosr2/uptool/commit/9c4be345745a07aae580477d15a5f408fa0d195c))
- Add signed commits check by [@santosr2](https://github.com/santosr2) ([87fb49d](https://github.com/santosr2/uptool/commit/87fb49dbe06934d21770be28f46dcb66987e7de5))
- Add security scan by [@santosr2](https://github.com/santosr2) ([9329007](https://github.com/santosr2/uptool/commit/9329007f86fe381df32440b34bb9b4dea26323d6))
- Add consolidate security workflow and workflow concurrency group by [@santosr2](https://github.com/santosr2) ([a0074b6](https://github.com/santosr2/uptool/commit/a0074b6741bc3f4746e0462dd1466c01f9f6ebd9))
- Fix security and docker workflows by [@santosr2](https://github.com/santosr2) ([979b394](https://github.com/santosr2/uptool/commit/979b39443d98e05d0f0fefa21c51b81628d5e884))
- Split scorecard due its requirements by [@santosr2](https://github.com/santosr2) ([b496a1f](https://github.com/santosr2/uptool/commit/b496a1f5513cff8a971c9f03c931a9065b07f176))
- Osv-scanner ignore more files by [@santosr2](https://github.com/santosr2) ([507ca3c](https://github.com/santosr2/uptool/commit/507ca3c185d33053542fd0b94de8cabd2539e8e1))
- Fix osv-scanner ignore files by [@santosr2](https://github.com/santosr2) ([ddae310](https://github.com/santosr2/uptool/commit/ddae3103e2f44b922a96c5baa60cf4542bff3da1))
- Fix osv-scanner lockfile for example/plugins by [@santosr2](https://github.com/santosr2) ([8540925](https://github.com/santosr2/uptool/commit/8540925da24589990d5b7059256e7bbe29aa95e4))
- Add pr workflows by [@santosr2](https://github.com/santosr2) ([38011f8](https://github.com/santosr2/uptool/commit/38011f8abb5633843527da129ff632e02c6c59f9))
- Add uptool action validation workflow by [@santosr2](https://github.com/santosr2) ([be61647](https://github.com/santosr2/uptool/commit/be61647a8fb1ff4773eb919bacbf265d2f8e65a0))
- Add license compliance workflow by [@santosr2](https://github.com/santosr2) ([d065b4e](https://github.com/santosr2/uptool/commit/d065b4e5b4cefc102b10a6eddf64b88666c97c12))
- Improve action validation workflow by [@santosr2](https://github.com/santosr2) ([9658427](https://github.com/santosr2/uptool/commit/9658427f7faeb38a36281604905c027fa51badd2))
- Add release pipelines by [@santosr2](https://github.com/santosr2) ([3c4a0f4](https://github.com/santosr2/uptool/commit/3c4a0f429910c73c3200b1843fb8b49efb44879b))
- Fix changelog workflow by [@santosr2](https://github.com/santosr2) ([cf2ea67](https://github.com/santosr2/uptool/commit/cf2ea67a2e131af41398e72218036fb7eee10c90))
- Fix bumpversion configuration by [@santosr2](https://github.com/santosr2) ([177a30e](https://github.com/santosr2/uptool/commit/177a30e813902e2c1fcdd7cc3c62a8aaadf4fe94))
- Fix changelog file by [@santosr2](https://github.com/santosr2) ([b48d805](https://github.com/santosr2/uptool/commit/b48d805a32e1d72e49166fd1a09dad8e58acfb40))
- Fix pre-release workflow by [@santosr2](https://github.com/santosr2) ([615df0c](https://github.com/santosr2/uptool/commit/615df0ce5f767873147574145c5bbe4255105045))
- Fix pre-release workflow by [@santosr2](https://github.com/santosr2) ([c83d83e](https://github.com/santosr2/uptool/commit/c83d83e4c623893d6a68fd673bf73f85fcbb4ef4))
- Improve pre-release commit logic by [@santosr2](https://github.com/santosr2) ([e6f0123](https://github.com/santosr2/uptool/commit/e6f01230951b310e36b51f57d871de1a44ffa737))
- Ensure commits are signed by [@santosr2](https://github.com/santosr2) ([3ce787c](https://github.com/santosr2/uptool/commit/3ce787ccc1d81e10383dfbc7d9fcf13c4aa09175))
- Update github-action[bot] email by [@santosr2](https://github.com/santosr2) ([738c495](https://github.com/santosr2/uptool/commit/738c495eeb7d4111cf0fcfbbd26f43db734bc0e8))
- Fix github-actions[bot] email by [@santosr2](https://github.com/santosr2) ([849f875](https://github.com/santosr2/uptool/commit/849f87543576b12360b5103b7d9a76f061577001))
- Fix workflow linter by [@santosr2](https://github.com/santosr2) ([b3cb868](https://github.com/santosr2/uptool/commit/b3cb868a1d5365a0aa0d4d35319f5406791529ac))

### Documentation
- Add community documentation by [@santosr2](https://github.com/santosr2) ([a633250](https://github.com/santosr2/uptool/commit/a6332506550814a8ec0534eac54e7cf7fba4a4fc))
- Use mkdocs instead of manual approach by [@santosr2](https://github.com/santosr2) ([248b2e0](https://github.com/santosr2/uptool/commit/248b2e0551e6ad73b23f6e7488a76c731830abd3))
- Minimize documentation verbosity by [@santosr2](https://github.com/santosr2) ([be57c28](https://github.com/santosr2/uptool/commit/be57c28bc5cf8886e3a0ca7388df9d0b7919ef13))

### Features
- Init repository by [@santosr2](https://github.com/santosr2) ([ad9be8c](https://github.com/santosr2/uptool/commit/ad9be8c63e430d119948a886424c28a689ab17ac))
- Add initial code structure by [@santosr2](https://github.com/santosr2) ([7ab22c5](https://github.com/santosr2/uptool/commit/7ab22c5db51fae4ec2e8a086120a2be1b6c66b3e))
- Add docs, examples and config files and improve code by [@santosr2](https://github.com/santosr2) ([e3da30d](https://github.com/santosr2/uptool/commit/e3da30dd0bd25874e687ff494fc2ed0e85f07e13))
- Add issues template by [@santosr2](https://github.com/santosr2) ([6829b5a](https://github.com/santosr2/uptool/commit/6829b5a8c8c98f44a30dc0816759dcc8ee25feda))
- Add cache to setup-mise action by [@santosr2](https://github.com/santosr2) ([2dfcaf9](https://github.com/santosr2/uptool/commit/2dfcaf9b96473ab2c80375a43bbb79454232a93c))
- Add container support by [@santosr2](https://github.com/santosr2) ([44fce45](https://github.com/santosr2/uptool/commit/44fce4576f8589567bfb2a8c081edbe744b78b34))

### GitHub Actions

- **build-release**: Remove unnecessary shell fields by [@santosr2](https://github.com/santosr2)([f6f48f6](https://github.com/santosr2/uptool/commit/f6f48f6306e3d1f429412dc396a37a67c714ddd5))

- **build-release**: Bump sbom-action to v0.20.10 by [@santosr2](https://github.com/santosr2)([b8513ad](https://github.com/santosr2/uptool/commit/b8513ad7234101619764b1f9d8ad680d47454202))

- **build-release**: Bump upload-artifact to v5.0.0 by [@santosr2](https://github.com/santosr2)([0b330fe](https://github.com/santosr2/uptool/commit/0b330fe65017ea8ad6069d986d029fb14b2cfb46))
- Ensure it will execute in proper folder by [@santosr2](https://github.com/santosr2) ([00056a9](https://github.com/santosr2/uptool/commit/00056a92368ee7f2c32682c06c264a829b18b392))
- Fix exit code by [@santosr2](https://github.com/santosr2) ([6cde855](https://github.com/santosr2/uptool/commit/6cde855566e6d78e138f98c4e6a99e4a56928edc))

### Miscellaneous Tasks

- **release**: Bump version to v0.2.0-alpha20251123 [skip ci] by [@github-actions[bot]](https://github.com/github-actions[bot])([fec28f6](https://github.com/santosr2/uptool/commit/fec28f6c448026d8a72c0454cdb31d1b80f41fdd))
- Remove mkdocs-material from mise due to an issue by [@santosr2](https://github.com/santosr2) ([5052509](https://github.com/santosr2/uptool/commit/5052509f7de1f371650749d85c58c3e20a46cc39))
- Migrate Makefile to mise tasks by [@santosr2](https://github.com/santosr2) ([9d12d54](https://github.com/santosr2/uptool/commit/9d12d5400f7f84122ab0ec7648d29318723d03d0))
- Fix docs and code based on what the linter finds by [@santosr2](https://github.com/santosr2) ([6a9f45e](https://github.com/santosr2/uptool/commit/6a9f45e2539c9ecfab90f84f4881fd28e5a8da33))
- Add PR and issue templates by [@santosr2](https://github.com/santosr2) ([da76029](https://github.com/santosr2/uptool/commit/da7602915bd87b1166b11da96b989f450afed9f4))
- Ensure go1.25 usage by [@santosr2](https://github.com/santosr2) ([d988ed6](https://github.com/santosr2/uptool/commit/d988ed6c2e478f5bc03b80d846bbfa360b377a66))
- Add codecov config by [@santosr2](https://github.com/santosr2) ([4511170](https://github.com/santosr2/uptool/commit/4511170e3a89d338c7bfd36101b2de4b02153c60))
- Fix linters by [@santosr2](https://github.com/santosr2) ([c604989](https://github.com/santosr2/uptool/commit/c60498969da041d21480f1d5a8b5fd54cbfc7e21))
- Consolidate security workflow and fix linters by [@santosr2](https://github.com/santosr2) ([a700bcd](https://github.com/santosr2/uptool/commit/a700bcda4d447ff33140115ef702476a7d0f1816))
- Add missing docs files by [@santosr2](https://github.com/santosr2) ([e619800](https://github.com/santosr2/uptool/commit/e6198003e0d6bdf2ffeb255ce449b16492c6e93b))
- Apply DRY in ValidateFilePath by [@santosr2](https://github.com/santosr2) ([3bd7b1c](https://github.com/santosr2/uptool/commit/3bd7b1cf78e4a936c1b643fb030e3e3fd4a332a1))
- License headers by [@santosr2](https://github.com/santosr2) ([b56a476](https://github.com/santosr2/uptool/commit/b56a4763287adf73b4e0bc4f671fe8b2417ce764))
- Remove .bumpversion.toml repeated logic by [@santosr2](https://github.com/santosr2) ([053d1cf](https://github.com/santosr2/uptool/commit/053d1cf8100f29a15e61d8f88826543cd8b58b11))
- : disable CGO in mise by [@santosr2](https://github.com/santosr2) ([5d1fdf4](https://github.com/santosr2/uptool/commit/5d1fdf47ba7489e9e0de79256ef435364ebce2a8))
- Improve bump version configuration by [@santosr2](https://github.com/santosr2) ([c1e42e1](https://github.com/santosr2/uptool/commit/c1e42e1ca971945114c34daf96fddc4e00589e72))

<!-- generated by git-cliff -->
