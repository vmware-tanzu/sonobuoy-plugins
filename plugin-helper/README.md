# Plugin Helper

A library made to trivialize making your containerized app into a Sonobuoy plugin.

## In Scope

- Submits results to the Sonobuoy aggregator
  - Can submit your own generated data (junit tests or general data)
  - Can write test results files for you
- Submits progress updates to the aggregator
  - Defaults to submitting a progress update for each test added

## Not In Scope

- This is not trying to be a generalized test runner or test framework. If you want that, there are lots of options which then can utilize this tool to submit the pre-generated data to the aggregator and even easily send updates as tests progress.