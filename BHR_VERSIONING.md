# BHRTech goja fork: branches and versioning

This repository is BHRTech's fork of [dop251/goja](https://github.com/dop251/goja).
It carries a small set of BHRTech patches on top of upstream and publishes
dated release tags that our projects pin via a `replace` directive.

## Branches

| Branch            | Content                                                        |
|-------------------|----------------------------------------------------------------|
| `master`          | Exact mirror of upstream `dop251/goja` `master`. Fast-forward only. Never commit to it. |
| `bhrtech-patches` | Upstream plus our patches. All releases are tagged on this branch. |

Upstream changes reach `bhrtech-patches` by **merging** `master` into it.
`bhrtech-patches` is never rebased and never force-pushed. Every commit that
was ever published stays reachable, so existing pins (including pseudo-versions
from before tags were introduced) keep resolving for `go mod download`,
`go mod tidy` and `go mod verify`.

To list our own patches: `git log --no-merges master..bhrtech-patches`

Mark patched code with a `// BHRTech patch:` comment explaining why the patch
exists, so it is easy to find during merge conflicts. Put tests for patches in
`bhrtech_patches_test.go` (run them with `go test -run BHRTech .`).

## Release tags

Format: `v0.0.0-bhrtech.YYYYMMDD.N`

- `YYYYMMDD` is the date the tag is created.
- `N` starts at `1` and increases for each further release on the same day
  (`v0.0.0-bhrtech.20261112.1`, `v0.0.0-bhrtech.20261112.2`, ...).
- Tags are only created on `bhrtech-patches`, and only after tests pass.

Why this shape:

- Go accepts only tags that are valid semver and start with `v` as module
  versions. A plain tag like `20261112.1` is treated as an arbitrary revision
  and ends up in `go.mod` as an unreadable pseudo-version.
- The date and sequence are numeric semver prerelease identifiers, so they
  compare numerically: later dates and higher sequence numbers always sort
  higher.
- The `bhrtech` identifier marks the version as a fork build. Because it is
  alphanumeric, it also sorts above the commit-hash pseudo-versions
  (`v0.0.0-<timestamp>-<hash>`) used before this scheme, so moving to a tag is
  an upgrade.
- `v0.0.0-...` cannot collide with any tag upstream might publish.

**Tags are immutable.** Never move or delete a tag after pushing it.
`proxy.golang.org` and `sum.golang.org` permanently record the first content
they see for a version, so a moved tag causes checksum errors for everyone. If
a release is broken, fix it and publish the next sequence number.

## One-time setup

```sh
git remote add upstream https://github.com/dop251/goja.git
git remote set-url --push upstream DISABLE
```

## Updating from upstream and releasing

```sh
git fetch upstream

# 1. Mirror upstream on master
git checkout master
git merge --ff-only upstream/master
git push origin master

# 2. Merge into the patch branch
git checkout bhrtech-patches
git merge master
#    resolve conflicts, then:
go vet ./...
go test ./...

# 3. Tag and publish
git tag v0.0.0-bhrtech.YYYYMMDD.N
git push origin bhrtech-patches v0.0.0-bhrtech.YYYYMMDD.N
```

## Using a release in a consuming project

The fork keeps the module path `github.com/dop251/goja`. Consumers keep
importing `github.com/dop251/goja` and redirect it with `replace`:

```sh
go mod edit -replace github.com/dop251/goja=github.com/bhrtech/goja@v0.0.0-bhrtech.YYYYMMDD.N
go mod tidy
go mod vendor   # if the project vendors its dependencies
```

`go get github.com/bhrtech/goja@...` does not work, because the fork's `go.mod`
declares `github.com/dop251/goja`.

If upstream raises its `go` directive, `go mod tidy` raises the consumer's
`go` directive to match.

If our patches are ever accepted upstream, dropping the fork means just
removing the `replace` line.
