package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	filename    = filepath.Base(os.Args[0])
	showVersion bool
	showHelp    bool
	version     = "dev"
	dryRun      bool
	inputDir    string
	jpgQuality  int
	pngQuality  int
	ignoreDirs  = []string{"www"}
	outputPath  = filename + ".txt"
	compressed  = make(map[string]struct{})
)

type image struct {
	path string
	size int64
	MD5  string
}

func init() {
	flag.BoolVar(&showVersion, "version", false, "print version number")
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.BoolVar(&dryRun, "dryrun", false, "run command without making changes")
	flag.StringVar(&inputDir, "input-dir", "", "the directory containing images to compress")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (runtime: %s)\n", filename, version, runtime.Version())
		os.Exit(0)
	}

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if inputDir == "" {
		fmt.Println("error: -input-dir is required")
		flag.Usage()
		os.Exit(2)
	}

	info, err := os.Stat(inputDir)
	if os.IsNotExist(err) {
		fmt.Println("error: path does not exist")
		os.Exit(2)
	}
	if !info.IsDir() {
		fmt.Println("error: specified path is not a directory")
		os.Exit(2)
	}

	loadCompressedMap()
	walkInputDir()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS]\n", filename)
	fmt.Fprintln(os.Stderr, "\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "\nEXAMPLES:")
	fmt.Fprintf(os.Stderr, "  %s -input-dir images\n", filename)
	fmt.Fprintf(os.Stderr, "  %s -input-dir images -dryrun\n", filename)
	fmt.Fprintln(os.Stderr, "")
}

func loadCompressedMap() {
	file, err := os.Open(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		compressed[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func walkInputDir() {
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range ignoreDirs {
				if d == dir {
					return filepath.SkipDir
				}
			}
		}
		// search path for images
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".png" {
			compress(path, ext, info.Size())
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func compress(path, ext string, size int64) {
	name := filepath.Base(path)

	// get MD5 of image
	fileMD5, err := getFileMD5(path)
	if err != nil {
		log.Fatalf("error: failed to get MD5 of file: %s", name)
	}

	// skip if already compressed
	if _, ok := compressed[fileMD5]; ok {
		return
	}

	// if not then compress and append new MD5 to file
	if dryRun {
		fmt.Printf("(dryrun) compressed: %s\n", path)
		return
	}

	switch ext {
	case ".jpg":
		guetzli(path)
	case ".png":
		zopflipng(path)
	default:
		log.Fatal("error: file is not an image")
	}

	newMD5, err := getFileMD5(path)
	if err != nil {
		log.Fatalf("error: failed to get MD5 of file: %s", name)
	}

	fi, err := os.Stat(path)
	if err != nil {
		log.Fatalf("error: failed to get size of compressed image: %s", err)
	}

	prevSize := byteCountIEC(size)
	newSize := byteCountIEC(fi.Size())
	fmt.Printf("compressed: %s from: %s to: %s\n", name, prevSize, newSize)
	writeMD5toFile(newMD5)
}

func guetzli(path string) {
	cmd := exec.Command("guetzli", "--quality", "84", path, path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = cmd.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error: compressing image: %s\n", out.String())
		log.Fatal(err)
	}
}

func zopflipng(path string) {
	cmd := exec.Command("zopflipng", "-m", "-y", path, path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = cmd.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error: compressing image: %s\n", out.String())
		log.Fatal(err)
	}
}

func getFileMD5(path string) (string, error) {
	var MD5 string
	file, err := os.Open(path)
	if err != nil {
		return MD5, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return MD5, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	MD5 = hex.EncodeToString(hashInBytes)
	return MD5, nil
}

func writeMD5toFile(fileMD5 string) {
	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(fileMD5 + "\n"); err != nil {
		log.Println(err)
	}
}

// convert a size in bytes to a human-readable string IEC (binary) format
// credit: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
