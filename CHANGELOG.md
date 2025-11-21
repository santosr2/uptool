# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes
- Ci, go.mod and golangci by @santosr2([4c3b789](https://github.com/santosr2/uptool/commit/4c3b7894fbc7d1158ba9f04c97366c983073f406))
- Remove docs symlinks since it breaks action cache by @santosr2([f2e1ea4](https://github.com/santosr2/uptool/commit/f2e1ea4bb116944d1e76f1de4cbbc943d0435984))
- Linters and CI by @santosr2([4a144d1](https://github.com/santosr2/uptool/commit/4a144d1c482d191a6f8cd3667769650939fc313a))

### Container
- Improve based on linter by @santosr2([3ffa920](https://github.com/santosr2/uptool/commit/3ffa92049e42b9d95e327c72e1b82f36e44fafb4))

### Continuous Integration
- Add ci pipeline by @santosr2([812bb4d](https://github.com/santosr2/uptool/commit/812bb4dc9ecb356a77db855b08cf7c11ab28793b))
- Add doc deploy by @santosr2([bc494c8](https://github.com/santosr2/uptool/commit/bc494c82dddf204a6397eeecdef0ea8ce6f06145))
- Fix doc workflow by @santosr2([7eeced3](https://github.com/santosr2/uptool/commit/7eeced3d13855caff9b5fd47a83c8f10fa169235))
- Fix golangci-lint by @santosr2([6cd077b](https://github.com/santosr2/uptool/commit/6cd077b18921b93f21a8f746538fa44c09caab41))
- Add signed commits check by @santosr2([f8227d5](https://github.com/santosr2/uptool/commit/f8227d538a90cf1e270266ebf86368de67269bc0))
- Add security scan by @santosr2([c3d0bed](https://github.com/santosr2/uptool/commit/c3d0bed34d689b8243bbb18b2666508ec9cc1723))
- Add consolidate security workflow and workflow concurrency group by @santosr2([545bf25](https://github.com/santosr2/uptool/commit/545bf2590588845a1106bafb2178c946116c2277))
- Fix security and docker workflows by @santosr2([978d2f0](https://github.com/santosr2/uptool/commit/978d2f0bed087248ef59f3ded7d9b59a3054bb6f))
- Split scorecard due its requirements by @santosr2([5fe23ce](https://github.com/santosr2/uptool/commit/5fe23ce86ceaf0ee31da70b95a44d2b01c281b80))
- Osv-scanner ignore more files by @santosr2([2177c5f](https://github.com/santosr2/uptool/commit/2177c5fca5657f5d98b51aa52ba815490ecaf336))
- Fix osv-scanner ignore files by @santosr2([3126d87](https://github.com/santosr2/uptool/commit/3126d872326afa322dae7dbacc0727299bb0f640))
- Fix osv-scanner lockfile for example/plugins by @santosr2([8c60aab](https://github.com/santosr2/uptool/commit/8c60aabac17b7efd8c39c122fd954cdc468e0f91))
- Add pr workflows by @santosr2([8f1052e](https://github.com/santosr2/uptool/commit/8f1052eae2d578d187bc5ca866ad516e98d1b87f))
- Add uptool action validation workflow by @santosr2([e360d70](https://github.com/santosr2/uptool/commit/e360d70d2071e7409da6490ad853d329e2ad7e36))
- Add license compliance workflow by @santosr2([fb2b031](https://github.com/santosr2/uptool/commit/fb2b0319fa761c073f7c7025cd68cc6782553f26))
- Improve action validation workflow by @santosr2([66a2de2](https://github.com/santosr2/uptool/commit/66a2de20b1eb9f5e023be39b7ba7333d3db52842))
- Add release pipelines by @santosr2([4a168a0](https://github.com/santosr2/uptool/commit/4a168a082999cae4730ca214a63004b37bd94996))
- Fix changelog workflow by @santosr2([6d2d596](https://github.com/santosr2/uptool/commit/6d2d596d28a2bfe2ffd2e4c940f8bcb5f29e809e))
- Fix bumpversion configuration by @santosr2([c6a691a](https://github.com/santosr2/uptool/commit/c6a691a0292e9afcc57dba8ca0c83638c1307ec8))
- Fix changelog file by @santosr2([efa5c82](https://github.com/santosr2/uptool/commit/efa5c82f4f1ab77cb92790583ab123bdf3257d8b))

### Documentation
- Update CHANGELOG.md [skip ci] by @github-actions[bot]([6c59527](https://github.com/santosr2/uptool/commit/6c595274d5b6324c629142762cf0f04bbfb20a96))
- Update CHANGELOG.md [skip ci] by @github-actions[bot]([5ac2e03](https://github.com/santosr2/uptool/commit/5ac2e03fd433ace4a44045caed527557482d5974))
- Update CHANGELOG.md [skip ci] by @github-actions[bot]([4e2a2b1](https://github.com/santosr2/uptool/commit/4e2a2b1772e8d69fbf4ea25508c15e457272bc4d))
- Update CHANGELOG.md [skip ci] by @github-actions[bot]([f127f83](https://github.com/santosr2/uptool/commit/f127f836a53a3945d86337a28469727e54299ab8))

### Features
- Add community documentation by @santosr2([7256083](https://github.com/santosr2/uptool/commit/7256083915b365c67603502d06f662874c9cedf6))
- Add docs, examples and config files and improve code by @santosr2([4c5e911](https://github.com/santosr2/uptool/commit/4c5e911dab21118922473408240f7aa7ffbf09d4))
- Use mkdocs instead of manual approach by @santosr2([459de5e](https://github.com/santosr2/uptool/commit/459de5e573be158fdd3caaec6d1697da1b73b92e))
- Add issues template by @santosr2([4a62df0](https://github.com/santosr2/uptool/commit/4a62df03e7fc2274ed7a27e5efc39e39f5a1d441))
- Add cache to setup-mise action by @santosr2([50d64f3](https://github.com/santosr2/uptool/commit/50d64f36039c34dc1f2bfcde773637f38928e9ac))
- Add container support by @santosr2([5f7ca50](https://github.com/santosr2/uptool/commit/5f7ca508b1637c4304e3da71bf66fab6cbcdaf53))

### GitHub Actions
- Ensure it will execute in proper folder by @santosr2([16f3c8c](https://github.com/santosr2/uptool/commit/16f3c8cb8e61460f5ca4d2e546432848a7e1cc67))
- Fix exit code by @santosr2([e17da0c](https://github.com/santosr2/uptool/commit/e17da0cd0ac12f210e50d144746afc0791f06d2a))

### Miscellaneous Tasks
- Remove mkdocs-material from mise due to an issue by @santosr2([cd7ac69](https://github.com/santosr2/uptool/commit/cd7ac69fdc778e3099370c58c8409b7d7f18c65c))
- Migrate Makefile to mise tasks by @santosr2([0187ac1](https://github.com/santosr2/uptool/commit/0187ac172d8b8eedc2d0f9739a145ba71c2d171a))
- Fix docs and code based on what the linter finds by @santosr2([7369b9a](https://github.com/santosr2/uptool/commit/7369b9ab0e61c641c972ef76cf31154c75f76246))
- Minimize documentation verbosity by @santosr2([61d0232](https://github.com/santosr2/uptool/commit/61d0232dd9efd618fc7c36c11e8aa2a61c99e291))
- Add PR and issue templates by @santosr2([5f51b52](https://github.com/santosr2/uptool/commit/5f51b52f384523bf6e71d9000d7e7c3de5ccf1ef))
- Ensure go1.25 usage by @santosr2([7ecded6](https://github.com/santosr2/uptool/commit/7ecded6db7d3a927a663bb4611dd5d367c74428f))
- Add codecov config by @santosr2([bfe4b5c](https://github.com/santosr2/uptool/commit/bfe4b5cf4b53826141b57866fea0e5c1f356f843))
- Fix linters by @santosr2([4b834da](https://github.com/santosr2/uptool/commit/4b834da993a587eeca0f77f9fef02a0e6253c5d5))
- Consolidate security workflow and fix linters by @santosr2([2486ab1](https://github.com/santosr2/uptool/commit/2486ab1cdc207c54ac16418f7110dcbd67d4f8ba))
- Add missing docs files by @santosr2([07f2c5f](https://github.com/santosr2/uptool/commit/07f2c5fe3bb0cbf1463e82d9761b8d206746c24e))
- Apply DRY in ValidateFilePath by @santosr2([a5478cf](https://github.com/santosr2/uptool/commit/a5478cfe03236c6e230a201edd2819b977acb9cd))
- License headers by @santosr2([25fb24c](https://github.com/santosr2/uptool/commit/25fb24c02b5c053848202980a399896a9bc196be))
- Remove .bumpversion.toml repeated logic by @santosr2([369e80b](https://github.com/santosr2/uptool/commit/369e80b6d4355469271b94e200d3a47bf4ae02fa))

## [0.1.0] - 2025-11-14

### Features
- Init repository by @santosr2([bbba065](https://github.com/santosr2/uptool/commit/bbba0652530ddd30e72e203df388e54b28502d3d))
- Add inital code structure by @santosr2([5eef376](https://github.com/santosr2/uptool/commit/5eef3764df1bec6b096d72cbf7ce06d90a1e34cf))

<!-- generated by git-cliff -->
