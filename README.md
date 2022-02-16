<p align="center"><a href="#readme"><img src="https://gh.kaos.st/check.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/g/check.v1"><img src="https://gh.kaos.st/godoc.svg" alt="PkgGoDev" /></a>
  <a href="https://kaos.sh/w/check/ci"><img src="https://kaos.sh/w/check/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/check/codeql"><img src="https://kaos.sh/w/check/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src="https://gh.kaos.st/mit.svg"></a>
</p>

The [Go language](http://golang.org) provides an internal testing library, named testing, which is relatively slim due to the fact that the standard library correctness by itself is verified using it. The check package, on the other hand, expects the standard library from Go to be working correctly, and builds on it to offer a richer testing framework for libraries and applications to use.

**This is fork of package [go-check/check](https://github.com/go-check/check) with some additional features.**

`gocheck` includes features such as:

* Helpful error reporting to aid on figuring problems out (_see below_)
* Richer test helpers: assertions which interrupt the test immediately, deep multi-type comparisons, string matching, etc
* Suite-based grouping of tests
* Fixtures: per suite and/or per test set up and tear down
* Benchmarks integrated in the suite logic (_with fixtures, etc_)
* Management of temporary directories
* Panic-catching logic, with proper error reporting
* Proper counting of successes, failures, panics, missed tests, skips, etc
* Explicit test skipping
* Support for expected failures
* Verbosity flag which disables output caching (_helpful to debug hanging tests, for instance_)
* Multi-line string reporting for more comprehensible failures
* Inclusion of comments surrounding checks on failure reports
* Fully tested (_it manages to test itself reliably_)

### Compatibility with `go test`

`gocheck` works as an extension to the testing package and to the "go test" runner. That allows keeping all current tests and using gocheck-based tests right away for new tests without conflicts. The gocheck API was purposefully made similar to the testing package for a smooth migration.

### Installing and updating

Install gocheck's check package with the following command:

```bash
go get -v pkg.re/essentialkaos/check.v1
```

To ensure you're using the latest version, run the following instead:

```bash
go get -u -v pkg.re/essentialkaos/check.v1
```

### API documentation

The API documentation for gocheck's check package is available online at:

https://pkg.re/essentialkaos/check.v1?docs

### Basic example

```golang
package hello_test

import (
    "testing"
    "io"

    . "pkg.re/essentialkaos/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestHelloWorld(c *C) {
    c.Assert(42, Equals, "42")
    c.Assert(io.ErrClosedPipe, ErrorMatches, "io: .*on closed pipe")
    c.Check(42, Equals, 42)
}
```

See [Assertions and checks](#assertions-and-checks) below for more information on these tests.

### Using fixtures

Fixtures are available by using one or more of the following methods in a test suite:

* `func (s *SuiteType) SetUpSuite(c *C)` - Run once when the suite starts running;
* `func (s *SuiteType) SetUpTest(c *C)` - Run before each test or benchmark starts running;
* `func (s *SuiteType) TearDownTest(c *C)` - Run after each test or benchmark runs;
* `func (s *SuiteType) TearDownSuite(c *C)` - Run once after all tests or benchmarks have finished running.

Here is an example preparing some data in a temporary directory before each test runs:

```golang
type Suite struct{
    dir string
}

func (s *MySuite) SetUpTest(c *C) {
    s.dir = c.MkDir()
    // Use s.dir to prepare some data.
}

func (s *MySuite) TestWithDir(c *C) {
    // Use the data in s.dir in the test.
}
```

### Adding benchmarks

Benchmarks may be added by prefixing a method in the suite with _Benchmark_. The method will be called with the usual `*C` argument, but unlike a normal test it is supposed to put the benchmarked logic within a loop iterating `c.N` times.

For example:

```golang
func (s *MySuite) BenchmarkLogic(c *C) {
    for i := 0; i < c.N; i++ {
        // Logic to benchmark
    }
}
```

These methods are only run when in benchmark mode, using the `-check.b` flag, and will present a result similar to the following when run:

```
PASS: myfile.go:67: MySuite.BenchmarkLogic       100000    14026 ns/op
PASS: myfile.go:73: MySuite.BenchmarkOtherLogic  100000    21133 ns/op
```

All the [fixture methods](#using-fixtures) are run as usual for a test method.

To obtain the timing for normal tests, use the `-check.v` flag instead.

### Skipping tests

Tests may be skipped with the [`Skip`](https://pkg.go.dev/pkg.re/essentialkaos/check.v1#C.Skip) method within `SetUpSuite`, `SetUpTest`, or the test method itself. This allows selectively ignoring tests based on custom factors such as the architecture being run, flags provided to the test, or the availbility of resources (_network, etc_).

As an example, the following test suite will skip all the tests within the suite unless the `-live` option is provided to `go test`:

```golang
var live = flag.Bool("live", false, "Include live tests")

type LiveSuite struct{}

func (s *LiveSuite) SetUpSuite(c *C) {
    if !*live {
        c.Skip("-live not provided")
    }
}
```

### Running tests and output sample

Use the `go test` tool as usual to run the tests:

```
$ go test

----------------------------------------------------------------------
FAIL: hello_test.go:16: S.TestHelloWorld

hello_test.go:17:
    c.Check(42, Equals, "42")
... obtained int = 42
... expected string = "42"

hello_test.go:18:
    c.Check(io.ErrClosedPipe, ErrorMatches, "BOOM")
... error string = "io: read/write on closed pipe"
... regex string = "BOOM"


OOPS: 0 passed, 1 FAILED
--- FAIL: hello_test.Test
FAIL
```

### Assertions and checks

gocheck uses two methods of `*C` to verify expectations on values obtained in test cases: [`Assert`](https://pkg.go.dev/pkg.re/essentialkaos/check.v1#C.Assert) and [`Check`](https://pkg.go.dev/pkg.re/essentialkaos/check.v1#C.Check). Both of these methods accept the same arguments, and the only difference between them is that when `Assert` fails, the test is interrupted immediately, while `Check` will fail the test, return false, and allow it to continue for further checks.

`Assert` and `Check` have the following types:

```golang
func (c *C) Assert(obtained interface{}, chk Checker, ...args interface{})
func (c *C) Check(obtained interface{}, chk Checker, ...args interface{}) bool
```

They may be used as follows:

```golang
func (s *S) TestSimpleChecks(c *C) {
    c.Assert(value, Equals, 42)
    c.Assert(s, Matches, "hel.*there")
    c.Assert(err, IsNil)
    c.Assert(foo, Equals, bar, Commentf("#CPUs == %d", runtime.NumCPU())
}
```

The last statement will display the provided message next to the usual debugging information, but only if the check fails.

Custom verifications may be defined by implementing the [Checker](https://pkg.go.dev/pkg.re/essentialkaos/check.v1#Checker) interface.

There are several standard checkers available:

**`DeepEquals`** — Сhecker verifies that the obtained value is deep-equal to the expected value. The check will work correctly even when facing slices, interfaces, and values of different types (_which always fail the test_).
```golang
c.Assert(array, DeepEquals, []string{"hi", "there"})
```

**`Equals`** — Checker verifies that the obtained value is equal to the expected value, according to usual Go semantics for `==`.
```golang
c.Assert(value, Equals, 42)
```

**`ErrorMatches`** — Checker verifies that the error value is non `nil` and matches the regular expression provided.
```golang
c.Assert(err, ErrorMatches, "perm.*denied")
```

**`FitsTypeOf`** — Checker verifies that the obtained value is assignable to a variable with the same type as the provided sample value.
```golang
c.Assert(value, FitsTypeOf, int64(0))
c.Assert(value, FitsTypeOf, os.Error(nil))
```

**`HasLen`** — Checker verifies that the obtained value has the provided length.
```golang
c.Assert(list, HasLen, 5)
```

**`Implements`** — Checker verifies that the obtained value implements the interface specified via a pointer to an interface variable.
```golang
var e os.Error
c.Assert(err, Implements, &e)
```

**`IsNil`** — Checker tests whether the obtained value is `nil`.
```golang
c.Assert(err, IsNil)
```

**`Matches`** — Checker verifies that the string provided as the obtained value (_or the string resulting from_ `obtained.String()`) matches the regular expression provided.
```golang
c.Assert(err, Matches, "perm.*denied")
```

**`NotNil`** — Checker verifies that the obtained value is not `nil`. This is an alias for `Not(IsNil)`, made available since it's a fairly common check.
```golang
c.Assert(iface, NotNil)
```

**`PanicMatches`** — Checker verifies that calling the provided zero-argument function will cause a panic with an error value matching the regular expression provided.
```golang
c.Assert(func() { f(1, 2) }, PanicMatches, `open.*: no such file or directory`)
```

**`Panics`** — Checker verifies that calling the provided zero-argument function will cause a panic which is deep-equal to the provided value.
```golang
c.Assert(func() { f(1, 2) }, Panics, &SomeErrorType{"BOOM"}).
```

### Selecting which tests to run

`gocheck` can filter tests out based on the test name, the suite name, or both. To run tests selectively, provide the command line option `-check.f` when running go test. Note that this option is specific to gocheck, and won't affect `go test` itself.

Some examples:

```
$ go test -check.f MyTestSuite
$ go test -check.f "Test.*Works"
$ go test -check.f "MyTestSuite.Test.*Works"
```

### Verbose modes

`gocheck` offers two levels of verbosity through the `-check.v` and `-check.vv` flags. In the first mode, passing tests will also be reported. The second mode will disable log caching entirely and will stream starting and ending suite calls and everything logged in between straight to the output. This is useful to debug hanging tests, for instance.

### Supported flags

- `check.f` - Regular expression selecting which tests and/or suites to run
- `check.v` - Verbose mode
- `check.vv` - Super verbose mode (*disables output caching*)
- `check.b` - Run benchmarks
- `check.btime` - Approximate run time for each benchmark (*default: 1 second*)
- `check.bmem` - Report memory benchmarks
- `check.list` - List the names of all tests that will be run
- `check.work` - Display and do not remove the test working directory
- `check.threads` - Number of parallel tests (*default: 1*)

### License

`gocheck` is made available under the [Simplified BSD License](LICENSE).
