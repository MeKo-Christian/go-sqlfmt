# Changelog

## [1.1.0](https://github.com/MeKo-Christian/go-sqlfmt/compare/v1.0.0...v1.1.0) (2025-12-12)


### Features

* added basic SQLite support ([4aed761](https://github.com/MeKo-Christian/go-sqlfmt/commit/4aed761375abce1806bff8132096e8b5dbe2ed21))
* added block context stack ([6d6e422](https://github.com/MeKo-Christian/go-sqlfmt/commit/6d6e4224bd576b9ecfe1e8a767e31a2ae035e230))
* Added Procedural Block Tracking to Formatter ([5ea92d0](https://github.com/MeKo-Christian/go-sqlfmt/commit/5ea92d08d77e320a9610e51ce1760f57ee2f9d6c))
* added tests ([ab6345b](https://github.com/MeKo-Christian/go-sqlfmt/commit/ab6345b1418c3a47b15fdd52727ddcec246582b6))
* added two options ([1638a47](https://github.com/MeKo-Christian/go-sqlfmt/commit/1638a47d57737083bb54bf924a4a207832fb36ab))
* Differentiate Statement Terminators from Query Separators ([7d4771b](https://github.com/MeKo-Christian/go-sqlfmt/commit/7d4771b0b992bd345cfea0fdbbd252cf47f8fc27))
* implement PostgreSQL support with numbered placeholders and related tests ([5ac9f9b](https://github.com/MeKo-Christian/go-sqlfmt/commit/5ac9f9bca7246ccc5f3fe9b0cb3e57930fcf3e43))
* major refactoring ([5e1f25b](https://github.com/MeKo-Christian/go-sqlfmt/commit/5e1f25b5c0b9140016f2fb6499f72df5361f0687))
* major refactoring ([7de9be2](https://github.com/MeKo-Christian/go-sqlfmt/commit/7de9be262884987c156aced14c7ea3b1fc7c1f8e))
* more dialects ([4a40d69](https://github.com/MeKo-Christian/go-sqlfmt/commit/4a40d695f4094245061b85482226491685844d85))
* more dialects ([23a25f5](https://github.com/MeKo-Christian/go-sqlfmt/commit/23a25f559a55927e1c307b86e5d8ae08ecd956d7))
* more enhancements ([72a11d5](https://github.com/MeKo-Christian/go-sqlfmt/commit/72a11d5bcae5cf2f1e82f817f7ed19d422412d03))
* more on indention ([5066e51](https://github.com/MeKo-Christian/go-sqlfmt/commit/5066e5194cb83451d94941b728b47dc2f83050fc))
* more qsqlite ([c5fb174](https://github.com/MeKo-Christian/go-sqlfmt/commit/c5fb174686152617ff633f93042f8682c4999b0b))
* more sql and documentation ([2a78390](https://github.com/MeKo-Christian/go-sqlfmt/commit/2a78390d78bb1cd01ee7c098485a51b04cf156dd))
* more sqlite ([7779996](https://github.com/MeKo-Christian/go-sqlfmt/commit/77799966b090152472e016266e9220956520ec3c))
* more sqlite ([83e248b](https://github.com/MeKo-Christian/go-sqlfmt/commit/83e248b772e6038fb51c38f5d8aa677501338bf5))
* more sqlite ([32ed8a3](https://github.com/MeKo-Christian/go-sqlfmt/commit/32ed8a3391964026eda6544ac004a0890e7b1fad))
* more sqlite ([bfc243a](https://github.com/MeKo-Christian/go-sqlfmt/commit/bfc243a9cd8d263bc2345bbe8d2027a9a1415f8d))
* Phase 2.5 done ([e0fa04c](https://github.com/MeKo-Christian/go-sqlfmt/commit/e0fa04cb236e110044ae5db0ef6b0beddd3f95c5))
* started on postgresql ([c1d1e52](https://github.com/MeKo-Christian/go-sqlfmt/commit/c1d1e52da60105b439640b6970af1ef9bcd84da2))


### Bug Fixes

* empty files should work without errors ([79983ae](https://github.com/MeKo-Christian/go-sqlfmt/commit/79983ae6ce19bb592b7502a88d7356bc0fd9bb66))
* failing tests ([45fa9ea](https://github.com/MeKo-Christian/go-sqlfmt/commit/45fa9ea61da6f16fdcec8b4df52e480f5a4de7be))
* formatting ([cef7c0c](https://github.com/MeKo-Christian/go-sqlfmt/commit/cef7c0ce9bccc963be70aaf2221dbecac5bef2c0))
* less lint errors ([60f41d0](https://github.com/MeKo-Christian/go-sqlfmt/commit/60f41d02189764640848f472ec763d7066cc6ee7))
* less lint stuff, more tests, but some still fail ([6dd76a0](https://github.com/MeKo-Christian/go-sqlfmt/commit/6dd76a04e6a39724842352adb012f506ad8049f4))
* lint errors ([6c8bdf8](https://github.com/MeKo-Christian/go-sqlfmt/commit/6c8bdf830963caea730c1633a05eb282618178c4))
* lint issues fixed ([db87c72](https://github.com/MeKo-Christian/go-sqlfmt/commit/db87c72ee329140879b9004c43a952da61842ee9))
* major refactoring ([4f91a5d](https://github.com/MeKo-Christian/go-sqlfmt/commit/4f91a5dfcbe7b8b437727c7d64110be125767671))
* more work ([28105a1](https://github.com/MeKo-Christian/go-sqlfmt/commit/28105a1d97f672776aef77c9e77ac9d19f8ffa4f))
* recent refinements ([a5e46e5](https://github.com/MeKo-Christian/go-sqlfmt/commit/a5e46e567c49bf8750841db18eae64b5ce43d1e9))
* semicolon issue fixed ([ac74858](https://github.com/MeKo-Christian/go-sqlfmt/commit/ac7485879fd29886ac895ebf403ae86b573e1b37))
* some more clean-up ([bdc80be](https://github.com/MeKo-Christian/go-sqlfmt/commit/bdc80be88788a7de5329e5fca8c67d71e3bca368))
* tests ([773f70f](https://github.com/MeKo-Christian/go-sqlfmt/commit/773f70fbc0140c0e04a4ea97cde52e9606d80e96))
* tests and more mysql ([a82dfc9](https://github.com/MeKo-Christian/go-sqlfmt/commit/a82dfc9707ed9e9342a05ca41936b086ab567794))
* tests for sqlite ([20ae676](https://github.com/MeKo-Christian/go-sqlfmt/commit/20ae676d407fb17dea78d1e7dfbe4cc96489cc46))
* version fix ([f20d4f1](https://github.com/MeKo-Christian/go-sqlfmt/commit/f20d4f1bef251286bba2e05bbb36fb838393033c))

## 1.0.0 (2025-12-11)

### Features

* Initial release with release-please automation
