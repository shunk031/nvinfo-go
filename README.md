nvinfo 
========

[![Release](https://github.com/shunk031/nvinfo-go/workflows/Release/badge.svg)](https://github.com/shunk031/nvinfo-go/actions?query=workflow%3ARelease) [![Latest Release](https://img.shields.io/github/release/shunk031/nvinfo-go.svg)](https://github.com/shunk031/nvinfo-go/releases)

Rewrite of [ikr7/nvinfo](https://github.com/ikr7/nvinfo) with Golang.

> nvinfo is a simple utility for monitoring your CUDA-enabled GPUs.

![](https://github.com/shunk031/nvinfo-go/raw/master/.github/screenshot.png)

## Usage

```sh
$ nvinfo
```

## Installation

Download the binary from [GitHub Releases](https://github.com/shunk031/nvinfo-go/releases/latest) and drop it in your `$PATH`.

- [Darwin / Mac](https://github.com/shunk031/nvinfo-go/releases/latest)
- [Linux](https://github.com/shunk031/nvinfo-go/releases/latest)

```sh
$ wget https://github.com/shunk031/nvinfo-go/releases/latest/download/nvinfo_linux_x86_64.tar.gz \
    && tar -xvzf nvinfo_linux_x86_64.tar.gz nvinfo \
    && rm nvinfo_linux_x86_64.tar.gz
```

## License

[MIT](https://github.com/shunk031/nvinfo-go/blob/master/LICENSE)
