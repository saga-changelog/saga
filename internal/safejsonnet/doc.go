// Package safejsonnet creates sandboxed go-jsonnet VMs for evaluating
// untrusted .jsonnet files (feats and config checked into a repo).
//
// The VMs are configured to limit resource consumption and prevent
// filesystem access beyond the .saga/ directory, making them safe to
// run in CI on pull requests from untrusted contributors.
package safejsonnet
