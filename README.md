# Dagger tests

## Install Dagger CLI

```
brew install dagger/tap/dagger
```


## Test one of the functions

```
dagger -m=github.com/rianfowler/dagger-module-demo call deploy-k-3-s-and-app terminal
```

Will put you in a terminal where you can type `kubectl get pods`

## Development

You must run for local development. It generates some files that are .gitignored by default

```
dagger develop --sdk=go
```