# DNS Checker

Given a collection of urls and expected redirects, validates and reports errors.

## Input Format

A CSV File with the source urls, the expected redirect and the expected status code (the first line is _always_ skipped):

```csv
"source";"target";"status"
"https://google.com";"https://www.google.com/";301
```

## Building

There's a Dockerfile and a GitHub action that creates releases for major OSes and architectures.
But if you just want to build a binary:

```shell
go build -o dns-checker .
```

## Command line

```shell
$ dns-checker --help
```

## Static x Dynamic check

Some redirects are dynamic (aka preserving the suffix: `foo.com/bar -> baz.com/bar`) others are static (aka always
redirect to the same destination: `foo.com/bar -> baz.com`).

We do the same kind of checking for now, but we have different flags for v.next.