# domaincheck

A fast CLI tool to check domain name availability using [RDAP](https://about.rdap.org/) (the successor to WHOIS).

## Features

- Concurrent domain lookups with configurable parallelism
- TLD expansion — check a name across multiple TLDs at once
- Read domains from arguments, a file, or stdin
- Colored terminal output
- Filter to show only available domains

## Install

```
go install github.com/AlexDobrushskiy/domaincheck@latest
```

Or build from source:

```
git clone https://github.com/AlexDobrushskiy/domaincheck.git
cd domaincheck
go build -o domaincheck .
```

## Usage

```
domaincheck example.com mysite.net coolproject.io
```

Check a name across multiple TLDs:

```
domaincheck -tlds com,net,io,dev myproject
```

Read from a file:

```
domaincheck -f domains.txt
```

Pipe from stdin:

```
echo -e "foo.com\nbar.net" | domaincheck
```

Show only available domains:

```
domaincheck -available-only -tlds com,net,io myproject
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f` | | Path to file with domains (one per line) |
| `-c` | `5` | Max concurrent checks |
| `-tlds` | | Comma-separated TLDs to expand base names |
| `-available-only` | `false` | Only print available domains |
| `-no-color` | `false` | Disable colored output |

## License

MIT
