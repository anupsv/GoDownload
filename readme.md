Certainly! Here's an updated `README.md` file that includes the new build options provided by the `Makefile`:

---

# GoDownload

GoDownload is a robust and concurrent file downloader written in Go. It allows users to download multiple files simultaneously, leveraging Go's goroutines for efficient and fast downloads. The downloader supports progress bars for each download, ensuring users are always informed about the status of their downloads.

## Features

- **Concurrent Downloads**: Download multiple files simultaneously.
- **Progress Bars**: Real-time progress bars for each download.
- **URL Validation**: Ensures only valid URLs are processed.
- **File Existence Check**: Skips downloading if the file already exists.
- **Graceful Exit**: Handles `CTRL+C` gracefully, ensuring all goroutines exit properly.

## Installation

To install GoDownload, you need to have Go installed on your machine. Once you have Go set up:

```bash
git clone https://github.com/yourusername/GoDownload.git
cd GoDownload
```

## Building

GoDownload provides a `Makefile` for easy building for various platforms:

- **macOS ARM**:
  ```bash
  make build-macos-arm
  ```

- **macOS Intel**:
  ```bash
  make build-macos-intel
  ```

- **Linux**:
  ```bash
  make build-linux
  ```

- **Windows**:
  ```bash
  make build-windows
  ```

- **Build for All Platforms**:
  ```bash
  make all
  ```

- **Clean Build Artifacts**:
  ```bash
  make clean
  ```

The binaries will be generated in the `build/` directory.

## Usage

### Command-Line Usage

```bash
./GoDownload -url https://example.com/file1.txt -url https://example.com/file2.jpg -dir /path/to/save
```

- `-url`: Specify the URL(s) to download. Can be used multiple times for multiple files.
- `-dir`: (Optional) Specify the directory where the files should be saved. Defaults to the current directory.
- `-threads`: (Optional) Specify the number of threads for downloading. Defaults to the number of CPUs.

### Examples

**Download a single file**:
```bash
./GoDownload -url https://example.com/file.txt
```

**Download multiple files**:
```bash
./GoDownload -url https://example.com/file1.txt -url https://example.com/file2.jpg
```

**Specify a directory to save the files**:
```bash
./GoDownload -url https://example.com/file.txt -dir /path/to/save
```

**Limit the number of threads**:
```bash
./GoDownload -url https://example.com/file.txt -threads 2
```

## Contributing

Contributions are welcome! Please fork the repository and create a pull request with your changes.

## License

This project is licensed under the MIT License.

---

This updated README now includes the build options and instructions on how to use the `Makefile` to build the project for various platforms.