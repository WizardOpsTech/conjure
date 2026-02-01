<!-- LOGO -->
<h1>
<p align="center">
  <img src="https://wizardops.dev/images/conjure-logo-white.svg" alt="Conjure Logo" width="128">
  <br>Conjure
</h1>
  <p align="center">
    Template-driven configuration generation for DevOps, Platform Engineers, and Developers.
    <br />
    <a href="#about">About</a>
    .
    <a href="https://github.com/wizardopstech/conjure/releases">Download</a>
    .
    <a href="https://conjure.wizardops.dev">Documentation</a>
    .
    <a href="#contributing">Contributing</a>
  </p>
</p>

## About

Conjure is a CLI tool that generates configurations from reusable templates.
Define a template once with Go template syntax, declare its variables in a
metadata file, and generate finished configs on demand. Conjure handles
Kubernetes manifests, Terraform modules, CI/CD pipelines, application configs,
and anything else that is text-based and follows a pattern.

Conjure is built for teams. Authors create templates and bundles that conform
to organizational standards. Consumers generate configurations by providing
variables through CLI flags, YAML values files, or interactive prompts --
no expertise with the underlying configuration format required.

For more details, see the [documentation](https://conjure.wizardops.dev).

## Download

See the [releases page](https://github.com/wizardopstech/conjure/releases)
on GitHub, or follow the [installation guide](https://conjure.wizardops.dev/docs/install/binary)
in the documentation.

Conjure ships as a single binary with no runtime dependencies.

## Documentation

See the [full documentation](https://conjure.wizardops.dev) on the Conjure website.

## Features and Status

|  #  | Feature                                           | Status |
| :-: | ------------------------------------------------- | :----: |
|  1  | Template generation with Go template syntax       |   Done |
|  2  | Bundle generation (multiple templates at once)    |   Done |
|  3  | Interactive mode with guided variable prompts     |   Done |
|  4  | Values files with variable precedence             |   Done |
|  5  | Local template and bundle repositories            |   Done |
|  6  | Remote repositories with SHA256 verification      |   Done |
|  7  | Repository index generation (`conjure repo index`)|   Done |
|  8  | Variable types: string, int, bool                 |   Done |

## Contributing

If you have ideas, issues, or would like to contribute to Conjure through
pull requests, please use the discussions to get started.

## License

Conjure is licensed under the [MIT License](LICENSE).
