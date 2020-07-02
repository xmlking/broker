# For Contributors

### Prerequisites

```bash
brew install hub
# goup checks if there are any updates for imports in your module.
GO111MODULE=on go get github.com/rvflash/goup
# for static check/linter
GO111MODULE=off go get github.com/golangci/golangci-lint/cmd/golangci-lint
```

### Test

```bash
make download
make test
make bench
# Benchmark specific test
cd pubsub
go test -run=^$ -bench=^BenchmarkInfoLog ./...
go test -run=^$ -bench=^BenchmarkInfoLog -benchtime 15s -count 2 -cpu 1,2,4 ./...
# Results:
# zerolog: 326443       3866 ns/op
# wrapper: 324487       4280 ns/op
```

### Release

```bash
make download
git add .
# Start release on develop branch
git flow release start v0.1.0
# on release branch
git-chglog -c .github/chglog/config.yml -o CHANGELOG.md --next-tag v0.1.0
# update `github.com/xmlking/broker` version in each `go.mod` file.
# commit all changes.
# finish release on release branch
git flow release finish
# on master branch, (gpoat = git push origin --all && git push origin --tags)
gpoat
# add git tags for sub-modules
make release TAG=v0.1.0
```

## Reference 
- https://github.com/nytimes/gizmo/blob/master/pubsub/gcp/gcp.go
- https://github.com/lileio/pubsub/blob/master/providers/google/google.go
- https://github.com/micro/go-plugins/tree/master/broker/googlepubsub
- https://github.com/cloudevents/sdk-go/blob/master/protocol/pubsub/v2/options.go
