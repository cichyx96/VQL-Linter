# VQL-Linter
Linter for VQL artifacts for Velociraptor.
Takes a yaml file or directory with yaml files and checks if they are valid VQL artifacts.

Return codes:
- 0: All files are valid VQL artifacts
- 1: Some files are not valid VQL artifacts


Usage
```
usage: vql-linter [<flags>] <target>

VQL linter for Velociraptor YAML artifacts.

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --disable-nested-lint  Disable linting of nested VQLs

Args:
  <target>  Path to yaml file or dir with yaml files to lint
```

Example
```
./vql-linter example_vqls/

- [bad.yaml] Failed to load YAML: While parsing source query: 1:10: unexpected token "=>" (expected "=" | "<=")
+ [ Custom.Example.Good ] Successfully compiled hunt
- [ Custom.Example.Nested.Bad ] Failed to compile hunt:  Unknown artifact reference Custom.Example.NONEXISTENT
+ [ Custom.Example.Nested.Good ] Successfully
```