# Overview
`img-compressor` is a command-line utility written in Go that compresses a directory of JPEG and PNG images using the [Zopfli PNG compressor][zopfli] and [Guetzli JPEG compressor][guetzli]. An MD5 of each compressed image will be written to a file (img-compressor.txt) so that on the next run, images that have already been compressed will be skipped.

## Building
Clone the repo then run the following commands:

```
go build
go install
```

To assign a version when building run:

```
go build -ldflags=-X=main.version=v1.0.0-beta1
```

## Dependencies
This program requires the following programs installed on your system.

- [Zopfli PNG compressor v1.0.3][zopfli]
- [Guetzli JPEG compressor v1.0.1][guetzli]

On macOS they can be installed using Homebrew with the following commands:

```
brew install zopfli
brew install guetzli
```

## Usage

```
img-compressor [OPTIONS]

OPTIONS:
  -dryrun
        run command without making changes
  -exclude string
        Glob pattern of directories/images to exclude, e.g {".git,*.jpg"}
  -help
        show help
  -input-dir string
        the directory containing images to compress
  -v    display more detailed output
  -version
        print version number


EXAMPLES:
  img-compressor -input-dir images
  img-compressor -input-dir images -dryrun
  img-compressor -input-dir . -exclude .git
  img-compressor -input-dir . -exclude {".git,*.jpg"}
```

## License
[MIT License](LICENSE)

[zopfli]: https://github.com/google/zopfli
[guetzli]: https://github.com/google/guetzli
