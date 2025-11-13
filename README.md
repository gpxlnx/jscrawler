## jscrawler

A high-performance tool for extracting JavaScript files from websites using concurrent crawling techniques.

## Installation

### Using Go Install
```
go install github.com/rix4uni/jscrawler@latest
```

### Download Prebuilt Binaries
```
wget https://github.com/rix4uni/jscrawler/releases/download/v0.0.4/jscrawler-linux-amd64-0.0.4.tgz
tar -xvzf jscrawler-linux-amd64-0.0.4.tgz
rm -rf jscrawler-linux-amd64-0.0.4.tgz
mv jscrawler ~/go/bin/jscrawler
```

Or download [binary release](https://github.com/rix4uni/jscrawler/releases) for your platform.

### Compile from Source
```
git clone --depth 1 https://github.com/rix4uni/jscrawler.git
cd jscrawler; go install
```

## Usage
```yaml
Usage of jscrawler:
      --complete        Get Complete URL (default false)
  -o, --output string   Output file to save results
      --silent          Silent mode.
  -t, --threads int     Number of threads to use (default 50)
      --timeout int     Timeout (in seconds) for http client (default 15)
      --verbose         Enable verbose output for debugging purposes.
      --version         Print the version of the tool and exit.
```

### Basic Syntax
```yaml
cat targets.txt | jscrawler [options]
```

or

```yaml
echo "https://example.com" | jscrawler [options]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--complete` | | Output complete URLs (absolute paths) | `false` |
| `--output` | `-o` | Save results to specified file | |
| `--silent` | | Suppress banner and non-essential output | `false` |
| `--threads` | `-t` | Number of concurrent threads | `50` |
| `--timeout` | | HTTP client timeout in seconds | `15` |
| `--verbose` | | Enable detailed debug output | `false` |
| `--version` | | Display version information and exit | |

## Examples

### Basic Usage
```yaml
# Single target
echo "https://example.com" | jscrawler

# Multiple targets from file
cat subdomains.txt | jscrawler
```

### Complete URL Extraction
```yaml
echo "https://example.com" | jscrawler --complete
```

### Save Results to File
```yaml
cat targets.txt | jscrawler --complete -o javascript_files.txt
```

### Advanced Operations
```yaml
# Verbose mode for debugging
echo "https://example.com" | jscrawler --verbose --complete

# Silent operation in pipelines
cat targets.txt | jscrawler --silent --complete

# Custom performance tuning
cat targets.txt | jscrawler --threads 100 --timeout 30
```

## Performance Comparison

jscrawler demonstrates superior coverage compared to similar tools:

```yaml
# Test against example.com
echo "https://www.example.com" | getJS --complete | wc -l
# Output: 4

echo "https://www.example.com" | subjs | wc -l  
# Output: 8

echo "https://www.example.com" | jscrawler --silent --complete | wc -l
# Output: 13
```

## Best Practices

### Input Preparation
```yaml
# Filter active hosts before processing
cat domains.txt | httpx -silent | jscrawler --complete
```

### Output Management
```yaml
# Remove duplicates and sort results
cat targets.txt | jscrawler --complete | unew | unique_js.txt
```

### Resource Optimization
```yaml
# Balance between speed and resource usage
cat large_target_list.txt | jscrawler --threads 50 --timeout 20
```