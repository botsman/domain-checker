# Domain Checker

A fast CLI tool to check domain availability across multiple combinations and TLDs.

## Features

- 🚀 Concurrent domain checking for speed
- 🔄 Two modes: single list combinations or cross-product of multiple lists
- 🌐 Support for multiple TLDs (.com, .net, .org, etc.)
- ✨ Flexible concatenation: with or without dash separator
- 📊 Clear console output with availability status

## Installation

```bash
# Clone the repository
cd domain-checker

# Install dependencies
go mod download

# Build the tool
go build -o domain-checker

# (Optional) Install to your system
go install
```

## Usage

### Basic Examples

**Check 2-word combinations from a single list:**
```bash
./domain-checker -keywords=super,fast,cloud
```
This checks: superfast.com, supercloud.com, fastcloud.com

**Check combinations between two lists:**
```bash
./domain-checker -lists="super,fast;cloud,service"
```
This checks: supercloud.com, superservice.com, fastcloud.com, fastservice.com

**Use dash separator:**
```bash
./domain-checker -keywords=my,app -dash
```
This checks: my-app.com

**Check multiple TLDs:**
```bash
./domain-checker -keywords=my,app -tlds=com,net,org
```
This checks: myapp.com, myapp.net, myapp.org

**Check 3-word combinations:**
```bash
./domain-checker -keywords=get,my,app,now -combinations=3
```
This checks: getmyapp.com, getmynow.com, getappnow.com, myappnow.com

### All Options

```
-keywords string
    Comma-separated keywords (e.g., 'one,two,three')
    
-lists string
    Semicolon-separated lists of keywords (e.g., 'one,two;three,four')
    Use this for cross-product mode instead of combinations
    
-combinations int
    Number of keywords to combine (default: 2)
    Ignored when -lists is provided
    
-tlds string
    Comma-separated TLDs to check (default: "com")
    Examples: "com,net,org" or "io,dev"
    
-dash
    Use dash separator (e.g., 'one-two' instead of 'onetwo')
    
-workers int
    Number of concurrent workers (default: 10)
    Increase for faster checking of large batches
```

## How It Works

1. **Single List Mode** (`-keywords`):
   - Takes one list of keywords
   - Generates all combinations of specified length (default 2)
   - Example: ["one", "two", "three"] with combinations=2 → ["onetwo", "onethree", "twothree"]

2. **List of Lists Mode** (`-lists`):
   - Takes multiple lists separated by semicolons
   - Generates cross-product between lists
   - Example: [["one", "two"], ["three", "four"]] → ["onethree", "onefour", "twothree", "twofour"]

3. **Domain Checking**:
   - Uses WHOIS protocol to check domain availability
   - Runs checks concurrently for speed
   - Automatically handles different WHOIS servers per TLD

## Output

The tool provides three sections:
- ✓ **AVAILABLE**: Domains available for registration
- ✗ **TAKEN**: Domains already registered
- ⚠ **ERRORS**: Domains that couldn't be checked (network issues, rate limiting, etc.)

## Examples with Real Domains

```bash
# Find available two-word domains
./domain-checker -keywords=quantum,cloud,fast,dev -tlds=com,io

# Find available startup names
./domain-checker -lists="get,my;started,going" -tlds=com,co

# Find available app names with dash
./domain-checker -keywords=todo,task,plan,track -dash -tlds=app,io
```

## Notes

- **Rate Limiting**: Some WHOIS servers may rate-limit requests. If you get errors, reduce the number of workers or add delays between batches.
- **Accuracy**: WHOIS responses vary by TLD. The tool uses common patterns to detect availability, but results should be verified.
- **Network**: Requires internet connection to query WHOIS servers.

## License

MIT
