# VQL-Linter
Linter for VQL artifacts for Velociraptor.
Takes a yaml file or directory with yaml files and checks if they are valid VQL artifacts.

If you need just simple linting of VQL files, you can use velociraptor binary with following command:
```
velociraptor --definitions /path/to/artifact.yaml artifacts show Custom.Artifact.Name
```

This will inform you about syntax errors in the VQL file, **however this won't check if all referenced artifacts exist and are valid**.

For comprehensive checks, you can use this linter.


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

Return codes:
- 0: All files are valid VQL artifacts
- 1: Some files are not valid VQL artifacts


Example
```
./vql-linter example_vqls/

- [bad.yaml] Failed to load YAML: While parsing source query: 1:10: unexpected token "=>" (expected "=" | "<=")
+ [ Custom.Example.Good ] Successfully compiled VQL
- [ Custom.Example.Nested.Bad ] Failed to compile VQL:  Unknown artifact reference Custom.Example.NONEXISTENT
+ [ Custom.Example.Nested.Good ] Successfully compiled VQL
```