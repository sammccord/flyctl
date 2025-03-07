# flyctl

flyctl is a command-line interface for fly.io

Note: Most installations of `flyctl` also alias `flyctl` to `fly` as a command name and this will become the default name in the future.
During the transition, note that where you see `flyctl` as a command it can be replaced with `fly`.

## Installation

## Using a Package Manager

#### [Homebrew](https://brew.sh) (macOS, Linux, WSL)

```bash
brew install superfly/tap/flyctl
```
To upgrade to the latest version:

```bash
brew upgrade flyctl
```

## Install Script

Download `flyctl` and install into 

Installing the latest version:

```bash
curl -L https://fly.io/install.sh | sh
```

Installing the latest pre-release version:

```bash
curl -L https://fly.io/install.sh | sh -s pre
```

Installing a specific version:

```bash
curl -L https://fly.io/install.sh | sh -s 0.0.200
```

## Downloading from GitHub

Download the appropriate version from the [Releases](https://github.com/sammccord/flyctl/releases) page of the `flyctl` GitHub repository.

## Getting Started

1. Sign into your fly account

```bash
flyctl auth login
```

2. List your apps

```bash
flyctl apps list
```

2. View app status

```bash
flyctl status -a {app-name}
```

## App Settings

`flyctl` will attempt to use the app name from a `fly.toml` file in the current directory. For example, if the current directory contains this file:


```bash
$ cat fly.toml
app: banana
```

`flyctl` will operate against the `banana` app unless overridden by the -a flag or other app name setting in the command line.

## Building on Windows

There is a simple Powershell script, `winbuild.ps1`, which will run the code generation for the help files, format them, and run a full build, leaving a new binary in the bin directory.

## Running from branches on your local machine

Run `scripts/build-dfly` to build a Docker image from the current branch. Then, use `scripts/dfly` to run it. This assumes you are already
authenticated to Fly in your local environment.

## Cutting a release

If you have write access to this repo, you can ship a prerelease or full release with:

`scripts/bump_version.sh prerel`

or

`scripts/bump_version.sh`
