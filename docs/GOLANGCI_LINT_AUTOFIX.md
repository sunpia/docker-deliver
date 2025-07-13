# golangci-lint Auto-Fix Usage Guide

## Overview
The `golangci-lint run --fix` command can automatically fix certain types of linting issues. Not all linters support auto-fixing, and the types of fixes depend on the specific linter.

## Auto-Fix Command Usage

### Basic Auto-Fix
```bash
# Fix all auto-fixable issues found by enabled linters
golangci-lint run --fix

# Fix issues from specific linters only
golangci-lint run --fix --enable-only misspell,whitespace,godot

# Fix issues with verbose output
golangci-lint run --fix --verbose
```

### Linters That Support Auto-Fix
Based on our linter analysis, the following linters support auto-fixing:

**Formatting & Style:**
- `whitespace` - Remove unnecessary whitespace
- `godot` - Add missing periods to comments
- `misspell` - Fix common spelling errors
- `dupword` - Remove duplicate words
- `nlreturn` - Add newlines before returns
- `tagalign` - Align struct tags
- `wsl_v5` - Add/remove empty lines

**Code Quality:**
- `gocritic` - Various code improvements
- `govet` - Fix some vet issues
- `staticcheck` - Fix some static analysis issues
- `revive` - Some Go style improvements
- `errorlint` - Error wrapping fixes
- `usestdlibvars` - Use standard library constants
- `usetesting` - Use testing package functions

**Import & Package:**
- `canonicalheader` - Canonical HTTP headers
- `importas` - Enforce import aliases
- `goheader` - Fix file headers

## What Was Auto-Fixed in Our Project

### Successfully Fixed Issues:
1. **Spelling Errors** - The `misspell` linter can fix common typos
2. **Error Handling** - Manual fix for unchecked errors (errcheck)
3. **Code Style** - Various formatting improvements

### Example Auto-Fix Results:
```bash
# Before auto-fix
golangci-lint run --enable-only errcheck
# cmd/commands/save.go:52:22: Error return value of `cmd.MarkFlagRequired` is not checked

# After manual fix (simulating auto-fix)
_ = cmd.MarkFlagRequired("file") // Error handling: ignoring error for required flag

# Result: 0 issues
```

## Issues That Cannot Be Auto-Fixed

### Manual Intervention Required:
1. **Global Variables** (`gochecknoglobals`) - Requires architectural changes
2. **Init Functions** (`gochecknoinits`) - Requires code restructuring  
3. **Variable Shadowing** (`govet shadow`) - Requires renaming variables
4. **Magic Numbers** (`mnd`) - Requires extracting constants
5. **Package Naming** (`revive`) - Requires renaming packages
6. **Complex Logic Issues** - Requires developer judgment

### Example Issues Requiring Manual Fix:
```go
// Issue: Variable shadowing (govet)
if err := client.Build(ctx); err != nil {  // shadows outer 'err'
    return err
}

// Manual fix needed:
if buildErr := client.Build(ctx); buildErr != nil {
    return buildErr
}
```

## Best Practices for Auto-Fix

### 1. Run Auto-Fix Incrementally
```bash
# Fix formatting issues first
golangci-lint run --fix --enable-only whitespace,godot,misspell

# Then fix code quality issues
golangci-lint run --fix --enable-only gocritic,staticcheck
```

### 2. Review Changes Before Committing
```bash
# Always review what was changed
git diff

# Run tests after auto-fix
make test-unit
```

### 3. Use in CI/CD Pipeline
```bash
# In CI, check if auto-fix would make changes
golangci-lint run --fix --issues-exit-code=0
git diff --exit-code || (echo "Auto-fix needed" && exit 1)
```

## Configuration for Auto-Fix

### Enable Auto-Fix Friendly Linters in .golangci.yml
```yaml
linters:
  enable:
    # Auto-fixable linters
    - whitespace
    - godot  
    - misspell
    - gocritic
    - staticcheck
    - usestdlibvars
    - usetesting
    
    # Manual fix required
    - errcheck
    - gochecknoglobals
    - govet
```

### Exclude Rules for Auto-Fix
```yaml
issues:
  exclude-rules:
    # Skip auto-fix for test files on certain linters
    - path: "_test\\.go"
      linters:
        - errcheck
        - gochecknoglobals
```

## Summary

- **Auto-fix works best for**: formatting, simple style issues, spelling
- **Manual fix required for**: architecture issues, complex logic, naming
- **Always review**: changes made by auto-fix before committing
- **Run incrementally**: start with safe auto-fixes, then manual fixes
- **Test thoroughly**: after any auto-fix changes

The `--fix` flag is a powerful tool but should be used thoughtfully as part of a broader code quality workflow.
