# Lopper

![icon](assets/icon.png)

[![CI](https://github.com/Piszmog/lopper/actions/workflows/ci.yml/badge.svg)](https://github.com/Piszmog/lopper/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Piszmog/lopper)](https://img.shields.io/github/v/release/Piszmog/lopper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Deletes dead local Git branches.

![running](assets/running.gif)

## Installation

Head over to [Releases](https://github.com/Piszmog/lopper/releases) and download the artifact for your architecture.

## Usage

```shell
$ ./lopper -p /path/to/repo/or/directory/of/repos 
```

### Options

| Option                 | Default | Required  | Description                                                          |
|:-----------------------|:-------:|:---------:|:---------------------------------------------------------------------|
| `--path`, `-p`         |   N/A   | **True**  | The path to the repository or directory of repositories              |
| `--protected-branches` |   N/A   | **False** | The branches other than `main` and `master` to protect from deletion |
| `--concurrency`, `-c`  |   `1`   | **False** | The number of repositories to process in parallel                    |
| `--dry-run`            | `false` | **False** | Run `lopper` without actually deleting any branches                  |
| `--help`, `-h`         | `false` | **False** | Shows help                                                           |

### Commands

| Command     | Description                                      |
|:------------|:-------------------------------------------------|
| `help`, `h` | Shows a list of commands or help for one command |
