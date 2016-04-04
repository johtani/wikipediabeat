# Wikipediabeat

Welcome to Wikipediabeat.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/johtani`

## Getting Started with Wikipediabeat

### Init Project
To get running with Wikipediabeat, run the following command:

```
make init
```

To commit the first version before you modify it, run:

```
make commit
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Wikipediabeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/johtani/wikipediabeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Wikipediabeat run the command below. This will generate a binary
in the same directory with the name wikipediabeat.

```
make
```


### Run

To run Wikipediabeat with debugging output enabled, run:

```
./wikipediabeat -c wikipediabeat.yml -e -d "*"
```


### Test

To test Wikipediabeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`


### Package

To cross-compile and package Wikipediabeat for all supported platforms, run the following commands:

```
cd dev-tools/packer
make deps
make images
make
```

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/wikipediabeat.template.json and etc/wikipediabeat.asciidoc

```
make update
```


### Cleanup

To clean  Wikipediabeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Wikipediabeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/johtani
cd ${GOPATH}/github.com/johtani
git clone https://github.com/johtani/wikipediabeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
