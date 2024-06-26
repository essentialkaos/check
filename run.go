package check

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

// -----------------------------------------------------------------------
// Test suite registry.

var allSuites []interface{}

// Suite registers the given value as a test suite to be run. Any methods
// starting with the Test prefix in the given value will be considered as
// a test method.
func Suite(suite interface{}) interface{} {
	allSuites = append(allSuites, suite)
	return suite
}

// -----------------------------------------------------------------------
// Public running interface.

var (
	filterFlag  = flag.String("check.f", "", "Regular expression selecting which tests and/or suites to run")
	verboseFlag = flag.Bool("check.v", false, "Verbose mode")
	streamFlag  = flag.Bool("check.vv", false, "Super verbose mode (disables output caching)")
	benchFlag   = flag.Bool("check.b", false, "Run benchmarks")
	benchTime   = flag.Duration("check.btime", 1*time.Second, "Approximate run time for each benchmark")
	benchMem    = flag.Bool("check.bmem", false, "Report memory benchmarks")
	listFlag    = flag.Bool("check.list", false, "List the names of all tests that will be run")
	workFlag    = flag.Bool("check.work", false, "Display and do not remove the test working directory")
	threadsNum  = flag.Int("check.threads", 1, "Number of parallel tests")
)

// TestingT runs all test suites registered with the Suite function,
// printing results to stdout, and reporting any failures back to
// the "testing" package.
func TestingT(testingT *testing.T) {
	conf := &RunConf{
		Filter:        *filterFlag,
		Verbose:       *verboseFlag || testing.Verbose(),
		Stream:        *streamFlag && (*threadsNum <= 1),
		Benchmark:     *benchFlag,
		BenchmarkTime: *benchTime,
		BenchmarkMem:  *benchMem,
		KeepWorkDir:   *workFlag,
	}
	if *listFlag {
		w := bufio.NewWriter(os.Stdout)
		for _, name := range ListAll(conf) {
			fmt.Fprintln(w, name)
		}
		w.Flush()
		return
	}
	result := RunAll(conf)
	println(result.String())
	if !result.Passed() {
		testingT.Fail()
	}
}

// RunAll runs all test suites registered with the Suite function, using the
// provided run configuration.
func RunAll(runConf *RunConf) *Result {
	result := Result{}
	queueCh := make(chan interface{})
	go func() {
		for _, s := range allSuites {
			queueCh <- s
		}
		close(queueCh)
	}()
	j := *threadsNum
	if j <= 1 {
		j = 1
	}
	resCh := make(chan *Result)
	for i := 0; i < j; i++ {
		go func() {
			for s := range queueCh {
				resCh <- Run(s, runConf)
			}
		}()
	}
	for i := 0; i < len(allSuites); i++ {
		result.Add(<-resCh)
	}
	return &result
}

// Run runs the provided test suite using the provided run configuration.
func Run(suite interface{}, runConf *RunConf) *Result {
	runner := newSuiteRunner(suite, runConf)
	return runner.run()
}

// ListAll returns the names of all the test functions registered with the
// Suite function that will be run with the provided run configuration.
func ListAll(runConf *RunConf) []string {
	var names []string
	for _, suite := range allSuites {
		names = append(names, List(suite, runConf)...)
	}
	return names
}

// List returns the names of the test functions in the given
// suite that will be run with the provided run configuration.
func List(suite interface{}, runConf *RunConf) []string {
	var names []string
	runner := newSuiteRunner(suite, runConf)
	for _, t := range runner.tests {
		names = append(names, t.String())
	}
	return names
}

// -----------------------------------------------------------------------
// Result methods.

func (r *Result) Add(other *Result) {
	r.Succeeded += other.Succeeded
	r.Skipped += other.Skipped
	r.Failed += other.Failed
	r.Panicked += other.Panicked
	r.FixturePanicked += other.FixturePanicked
	r.ExpectedFailures += other.ExpectedFailures
	r.Missed += other.Missed
	if r.WorkDir != "" && other.WorkDir != "" {
		r.WorkDir += ":" + other.WorkDir
	} else if other.WorkDir != "" {
		r.WorkDir = other.WorkDir
	}
}

func (r *Result) Passed() bool {
	return (r.Failed == 0 && r.Panicked == 0 &&
		r.FixturePanicked == 0 && r.Missed == 0 &&
		r.RunError == nil)
}

func (r *Result) String() string {
	if r.RunError != nil {
		return "ERROR: " + r.RunError.Error()
	}

	var value string
	if r.Failed == 0 && r.Panicked == 0 && r.FixturePanicked == 0 &&
		r.Missed == 0 {
		value = "OK: "
	} else {
		value = "OOPS: "
	}
	value += fmt.Sprintf("%d passed", r.Succeeded)
	if r.Skipped != 0 {
		value += fmt.Sprintf(", %d skipped", r.Skipped)
	}
	if r.ExpectedFailures != 0 {
		value += fmt.Sprintf(", %d expected failures", r.ExpectedFailures)
	}
	if r.Failed != 0 {
		value += fmt.Sprintf(", %d FAILED", r.Failed)
	}
	if r.Panicked != 0 {
		value += fmt.Sprintf(", %d PANICKED", r.Panicked)
	}
	if r.FixturePanicked != 0 {
		value += fmt.Sprintf(", %d FIXTURE-PANICKED", r.FixturePanicked)
	}
	if r.Missed != 0 {
		value += fmt.Sprintf(", %d MISSED", r.Missed)
	}
	if r.WorkDir != "" {
		value += "\nWORK=" + r.WorkDir
	}
	return value
}
