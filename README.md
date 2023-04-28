
### gh-qldb

Tired of having dozens of CodeQL databases scattered around your file system? Introducing QLDB, a CodeQL database manager. Download, deploy and create CodeQL databases with ease.

QLDB will organize your databases in a hierarchical structure:

```bash
/Users/pwntester/codeql-dbs
└── github.com
   ├── apache
   │  ├── logging-log4j2
   │  │  ├── java
   │  │  │  └── fa2f51eb8d5912b92b8aa75693464f3e55d36be9.zip
   │  │  └── javascript
   │  │     └── abf13fabc45a2f51eb8d5912b92b8aa75693464f.zip
   │  └── commons-text
   │     └── java
   │        └── e2b291e92bcc272f98d6ca6723994da985d30e07.zip
   └── pwntester
      └── sample-project
         └── java
            └── 9b84404246d516a11091e74ef4cdcf7dfcc63fa4.zip
```

### Usage

```bash
Usage:
  gh qldb [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Extracts a CodeQL database from a source path
  download    Downloads a CodeQL database from GitHub Code Scanning
  help        Help about any command
  install     Install a local CodeQL database in the QLDB directory
  info        Returns information about a database stored in the QLDB structure

Flags:
  -h, --help   help for gh-qldb
```

### Examples

#### Create a database

```bash
gh qldb create -n foo/bar -- -s path/to/src -l java
```

#### Download a Code Scanning database

```bash
gh qldb download -n apache/logging-log4j2 -l java
```

#### Install a local database in QLDB structure

```bash
gh qldb install -d path/to/database -n apache/logging-log4j2
```

#### Get information about a database

```bash
gh qldb info -n apache/logging-log4j2 -l java -j
[
  {
    "commitSha": "fa2f51e",
    "committedDate": "2023-04-06T06:25:30",
    "path": "/Users/pwntester/codeql-dbs/github.com/apache/logging-log4j2/java/9b84404246d516a11091e74ef4cdcf7dfcc63fa4.zip
  }
]
```

### Similar projects

Liked the idea? Do you want to use a similar functionality for managing your GitHub projects and clones? Try [`gh cdr`](https://github.com/pwntester/gh-cdr)
