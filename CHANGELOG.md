# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes
- Ci, go.mod and golangci by [@santosr2](https://github.com/santosr2) ([4d59315](https://github.com/santosr2/uptool/commit/4d59315e55438836d0be2e64d40ce561ea8bf328))
- Remove docs symlinks since it breaks action cache by [@santosr2](https://github.com/santosr2) ([5d22735](https://github.com/santosr2/uptool/commit/5d227353250e3fc2c3f99ee69c9dc64345328c7d))
- Linters and CI by [@santosr2](https://github.com/santosr2) ([162cec1](https://github.com/santosr2/uptool/commit/162cec15b6f9e23c1b690da0b4ec0a8352e445f6))

### Container
- Improve based on linter by [@santosr2](https://github.com/santosr2) ([d98ea80](https://github.com/santosr2/uptool/commit/d98ea803d99d424749b644ef124939b7777504ad))

### Continuous Integration

- **changelog**: Ensure actors are tagged by [@santosr2](https://github.com/santosr2)([d29b2d9](https://github.com/santosr2/uptool/commit/d29b2d91ad7eb3bac0813a6d591844b98a019a8d))

- **docker**: Fix scan command by [@santosr2](https://github.com/santosr2)([16601c1](https://github.com/santosr2/uptool/commit/16601c1f3b08f8853ffa3e958c1cb4f9eb11959f))

- **pre-release**: Remove unnecessary sed command by [@santosr2](https://github.com/santosr2)([e45e66c](https://github.com/santosr2/uptool/commit/e45e66c040ba201267593f795759ed665fbbcffd))

- **pre-release**: Fix checkout strategy by [@santosr2](https://github.com/santosr2)([45df0f4](https://github.com/santosr2/uptool/commit/45df0f4333e5fbb08122209506830b2e1778466b))

- **release**: Remove C dependency by [@santosr2](https://github.com/santosr2)([ee5756c](https://github.com/santosr2/uptool/commit/ee5756c03495206570b3e33a368391ed1ae9669b))

- **signed-commits**: Ensure it fails when no-signed commits are found by [@santosr2](https://github.com/santosr2)([20721d2](https://github.com/santosr2/uptool/commit/20721d26fef09fa34e81461c3d3095ef76b44d31))
- Add ci pipeline by [@santosr2](https://github.com/santosr2) ([11e31d5](https://github.com/santosr2/uptool/commit/11e31d5d828b4f83674235fc1b0f208204f88dc1))
- Add doc deploy by [@santosr2](https://github.com/santosr2) ([b2a3aea](https://github.com/santosr2/uptool/commit/b2a3aea0f40ba9a1c28f09822f8562455ca87bbc))
- Fix doc workflow by [@santosr2](https://github.com/santosr2) ([9dd9ede](https://github.com/santosr2/uptool/commit/9dd9ede78e7eb61078e46c8507165fc9a8b2d426))
- Fix golangci-lint by [@santosr2](https://github.com/santosr2) ([b89379f](https://github.com/santosr2/uptool/commit/b89379f01cb5b8bcf6f737f62f7698f50a784dc5))
- Add signed commits check by [@santosr2](https://github.com/santosr2) ([7923adb](https://github.com/santosr2/uptool/commit/7923adb892e703bfbaa5122383d2e9417fe93523))
- Add security scan by [@santosr2](https://github.com/santosr2) ([165a674](https://github.com/santosr2/uptool/commit/165a674beb6db1a462a3a249b1e47e1c6c644b58))
- Add consolidate security workflow and workflow concurrency group by [@santosr2](https://github.com/santosr2) ([8232f13](https://github.com/santosr2/uptool/commit/8232f13430743f73b69b38cf2994f6046cd6692a))
- Fix security and docker workflows by [@santosr2](https://github.com/santosr2) ([c7ef403](https://github.com/santosr2/uptool/commit/c7ef4032cc1035f7b8d33e251a9bc408078255fa))
- Split scorecard due its requirements by [@santosr2](https://github.com/santosr2) ([6ced958](https://github.com/santosr2/uptool/commit/6ced958639aab207ae9941d31a4c0e84c742bc0d))
- Osv-scanner ignore more files by [@santosr2](https://github.com/santosr2) ([eb6a1ca](https://github.com/santosr2/uptool/commit/eb6a1ca3eb34dd983a22ae70f74a3647dc1c6cbf))
- Fix osv-scanner ignore files by [@santosr2](https://github.com/santosr2) ([6cfa8f2](https://github.com/santosr2/uptool/commit/6cfa8f2ac74c31b5d679b98747d53783a4187a13))
- Fix osv-scanner lockfile for example/plugins by [@santosr2](https://github.com/santosr2) ([b76cfa9](https://github.com/santosr2/uptool/commit/b76cfa928d38c4da3a9a23f953c55c4798809d95))
- Add pr workflows by [@santosr2](https://github.com/santosr2) ([11ac31b](https://github.com/santosr2/uptool/commit/11ac31be8be6ed4e380a3a7fa29744941061e20a))
- Add uptool action validation workflow by [@santosr2](https://github.com/santosr2) ([7db2f8c](https://github.com/santosr2/uptool/commit/7db2f8cd81d0955c38bb4f829bdd804e4b93b669))
- Add license compliance workflow by [@santosr2](https://github.com/santosr2) ([4a54aba](https://github.com/santosr2/uptool/commit/4a54aba34a7a922d166467431de43c903c4d2b58))
- Improve action validation workflow by [@santosr2](https://github.com/santosr2) ([91ccbb0](https://github.com/santosr2/uptool/commit/91ccbb01acede029e926749725008059c18c3db8))
- Add release pipelines by [@santosr2](https://github.com/santosr2) ([637d34c](https://github.com/santosr2/uptool/commit/637d34c24bdfe360fa000641fd59cdc80c5e13a5))
- Fix changelog workflow by [@santosr2](https://github.com/santosr2) ([d8d6390](https://github.com/santosr2/uptool/commit/d8d6390e36e9aeed8ee88cc3de1b77a96fd5d098))
- Fix bumpversion configuration by [@santosr2](https://github.com/santosr2) ([82bbaf6](https://github.com/santosr2/uptool/commit/82bbaf6f1c71c108a4ba707402a44b7a30b8458d))
- Fix changelog file by [@santosr2](https://github.com/santosr2) ([0dad9da](https://github.com/santosr2/uptool/commit/0dad9dadc94f5f3b44cad3523e6bb2ea548b7141))
- Fix pre-release workflow by [@santosr2](https://github.com/santosr2) ([6082bae](https://github.com/santosr2/uptool/commit/6082bae2bf3b903b0916516bb9687b9421c62f0d))
- Fix pre-release workflow by [@santosr2](https://github.com/santosr2) ([c4b2f05](https://github.com/santosr2/uptool/commit/c4b2f05b1ce6daab374055334315b77d5aef03f7))
- Improve pre-release commit logic by [@santosr2](https://github.com/santosr2) ([50ae5e0](https://github.com/santosr2/uptool/commit/50ae5e06abbc35ca6e0d32d79638cbf5fc85cfda))
- Ensure commits are signed by [@santosr2](https://github.com/santosr2) ([d15b43a](https://github.com/santosr2/uptool/commit/d15b43af360e64f42c41803cf1a309f65c7db0e4))
- Update github-action[bot] email by [@santosr2](https://github.com/santosr2) ([4f2b70a](https://github.com/santosr2/uptool/commit/4f2b70a62563276fc814b4642c7c369801554784))
- Fix github-actions[bot] email by [@santosr2](https://github.com/santosr2) ([917a39c](https://github.com/santosr2/uptool/commit/917a39cc29af6943af643b353aa9ea9ce767085f))

### Features
- Init repository by [@santosr2](https://github.com/santosr2) ([ad9be8c](https://github.com/santosr2/uptool/commit/ad9be8c63e430d119948a886424c28a689ab17ac))
- Add inital code structure by [@santosr2](https://github.com/santosr2) ([029945a](https://github.com/santosr2/uptool/commit/029945a22a879c40fe45444511cfc4658d5bd2b8))
- Add community documentation by [@santosr2](https://github.com/santosr2) ([7f93dd0](https://github.com/santosr2/uptool/commit/7f93dd065d91ec55f65ac9f55f8cd27c9a11a20c))
- Add docs, examples and config files and improve code by [@santosr2](https://github.com/santosr2) ([9551b11](https://github.com/santosr2/uptool/commit/9551b115a24b27f8533a658235140f88e022814d))
- Use mkdocs instead of manual approach by [@santosr2](https://github.com/santosr2) ([b8b4673](https://github.com/santosr2/uptool/commit/b8b4673019a7c4a4ecf5b777a0259ca5a96d8b62))
- Add issues template by [@santosr2](https://github.com/santosr2) ([6633b65](https://github.com/santosr2/uptool/commit/6633b65f7dae11354fc532245ff31f7f4442d13f))
- Add cache to setup-mise action by [@santosr2](https://github.com/santosr2) ([ac9eac8](https://github.com/santosr2/uptool/commit/ac9eac8c1869cf331ae516d5a177873c6b7754ff))
- Add container support by [@santosr2](https://github.com/santosr2) ([963e1a6](https://github.com/santosr2/uptool/commit/963e1a65190b9b180d8938123bfc0ef4218111a8))

### GitHub Actions

- **build-release**: Remove unnecessary shell fields by [@santosr2](https://github.com/santosr2)([74b88da](https://github.com/santosr2/uptool/commit/74b88da62f362766e8d05e5331b44a4d7b5aa258))

- **build-release**: Bump sbom-action to v0.20.10 by [@santosr2](https://github.com/santosr2)([e7b7027](https://github.com/santosr2/uptool/commit/e7b7027a0b916e9dddb69f8cdb81c1e0b18c4528))

- **build-release**: Bump upload-artifact to v5.0.0 by [@santosr2](https://github.com/santosr2)([8569ab5](https://github.com/santosr2/uptool/commit/8569ab58a867889cd47cafb7a6d6477cb59b422e))
- Ensure it will execute in proper folder by [@santosr2](https://github.com/santosr2) ([3cb0acf](https://github.com/santosr2/uptool/commit/3cb0acfb3cdf6be2f424451ad921f285c43073fc))
- Fix exit code by [@santosr2](https://github.com/santosr2) ([91f67b9](https://github.com/santosr2/uptool/commit/91f67b9ac82d484f4b68c18dbe3d610d6255f954))

### Miscellaneous Tasks

- **release**: Bump version to v0.2.0-alpha20251123 [skip ci] by [@github-actions[bot]](https://github.com/github-actions[bot])([59d0f00](https://github.com/santosr2/uptool/commit/59d0f007a56c4ec9c8780dc4d68ce237dd0d2a80))
- Remove mkdocs-material from mise due to an issue by [@santosr2](https://github.com/santosr2) ([719c55d](https://github.com/santosr2/uptool/commit/719c55d90141ddd8ba6a3e6de9f272cf210c064f))
- Migrate Makefile to mise tasks by [@santosr2](https://github.com/santosr2) ([8242b91](https://github.com/santosr2/uptool/commit/8242b914dda97dd141e5d16e1382bbba454b6a1e))
- Fix docs and code based on what the linter finds by [@santosr2](https://github.com/santosr2) ([434acb0](https://github.com/santosr2/uptool/commit/434acb088230b47c7a15ab846a8c22fc9d048d3e))
- Minimize documentation verbosity by [@santosr2](https://github.com/santosr2) ([31fb9ba](https://github.com/santosr2/uptool/commit/31fb9bad0d090f36119814b7a13211fc11d315cd))
- Add PR and issue templates by [@santosr2](https://github.com/santosr2) ([eabdbc3](https://github.com/santosr2/uptool/commit/eabdbc38c113aba6926720203585977642a350c3))
- Ensure go1.25 usage by [@santosr2](https://github.com/santosr2) ([7f93941](https://github.com/santosr2/uptool/commit/7f939412ee7d459194cee4c7d4e6d4008dc5998e))
- Add codecov config by [@santosr2](https://github.com/santosr2) ([51aa338](https://github.com/santosr2/uptool/commit/51aa338fb3400250f829eef1ad015625494ee4b3))
- Fix linters by [@santosr2](https://github.com/santosr2) ([4350821](https://github.com/santosr2/uptool/commit/435082103efef0b94160b79fd21c2dc400802eda))
- Consolidate security workflow and fix linters by [@santosr2](https://github.com/santosr2) ([ff338f2](https://github.com/santosr2/uptool/commit/ff338f21a7856204333ddd923c5f34f9069303a3))
- Add missing docs files by [@santosr2](https://github.com/santosr2) ([c185e9c](https://github.com/santosr2/uptool/commit/c185e9c5e3b104e7c6db9284661ab589ffaea5eb))
- Apply DRY in ValidateFilePath by [@santosr2](https://github.com/santosr2) ([ef9c8d6](https://github.com/santosr2/uptool/commit/ef9c8d6da7003de4e7bca8e81219a1074b700467))
- License headers by [@santosr2](https://github.com/santosr2) ([9d47e2b](https://github.com/santosr2/uptool/commit/9d47e2bfb1e870a94db6eccea2f1eceef1747427))
- Remove .bumpversion.toml repeated logic by [@santosr2](https://github.com/santosr2) ([04f7277](https://github.com/santosr2/uptool/commit/04f72776927293393b380232030040d4a558bdae))
- : disable CGO in mise by [@santosr2](https://github.com/santosr2) ([4c71a8e](https://github.com/santosr2/uptool/commit/4c71a8e062c92804b47ced81d4d5966eb5ae9108))
- Improve bump version configuration by [@santosr2](https://github.com/santosr2) ([ca61588](https://github.com/santosr2/uptool/commit/ca6158893db7468a09838b136dcea7e3264cb38d))

<!-- generated by git-cliff -->
