gogocyclo reads output from github.com/fzipp/gocyclo and ignores results that
match any rules defined in a configuration file. Intended to be used
as a build step in CI (like Travis CI and CircleCI).

To install:

```
$ go get github.com/limouren/gogocyclo
```

Usage:

Use injunction with gocyclo

```
$ gocyclo -over 20 $GOPATH/src/database/sql | gogocylo -config .gogocyclo.sample
```

`-config` should point to a gogocyclo configuration. It defaults to `.gogocyclo`
if omitted. Here is a sample config for `database/sql` intended for the command
above.

```
[gogocyclo]
ignores = `sql`.convertAssign
ignores = `sql`.TestConversions
ignores = `sql`.TestMaxOpenConns
ignores = `sql`.(*fakeStmt).Query
```

See the program usage for details on the format of ignores:

```
$ gogocyclo -h
```
