# Scenario

## Setup — shared preconditions

A temporary home directory is created under `/tmp` for isolated testing.
The `HOME` env var is set in the `Run` function to point to this temp dir
so all path resolution uses the controlled value.

## Steps

1. Set `HOME` env var to the temp directory
2. Call the config path function under test
3. Verify the returned path matches expected `$HOME`-relative path
