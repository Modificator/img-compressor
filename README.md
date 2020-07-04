# Overview
`img-compressor` is a command-line utility written in Go that compresses a directory of JPEG and PNG images using the [Zopfli PNG compressor][zopfli] and [Guetzli JPEG compressor][guetzli]. An MD5 of each compressed image will be written to a file (compressed.txt) so that on the next run, images that have already been compressed will be skipped.

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
Compress all JPG and PNG images located in `source` and copy them to the`dest` directory.

```
img-compressor -input-dir source -output-dir dest
```

## License
[MIT License](LICENSE)

[zopfli]: https://github.com/google/zopfli
[guetzli]: https://github.com/google/guetzli
