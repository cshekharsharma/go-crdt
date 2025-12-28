#!/bin/bash

set -euo pipefail

print_start() {
  echo -e "\n\033[1;33m=============================="
  echo -e "  🚧 RUNNING PRE-COMMIT HOOK  "
  echo -e "==============================\033[0m\n"
}

print_success() {
  echo -e "\n\033[1;32m=============================="
  echo -e "  ✅ Pre-Commit hook passed."
  echo -e "==============================\033[0m\n"
}

print_failure() {
  echo -e "\n\033[1;31m=============================="
  echo -e "  ❌ Pre-Commit hook failed."
  echo -e "==============================\033[0m\n"
}

trap 'echo -e "\n💥 An unexpected error occurred. Aborting commit."; print_failure; exit 1' ERR

print_start

echo -e "\033[1;36m>> Checking staged Go files for commit...\033[0m\n"

staged_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -z "$staged_files" ]]; then
  echo -e "✅ No Go files staged for commit. Skipping checks.\n"
  print_success
  exit 0
fi

echo -e "Staged Go files:"
echo "$staged_files" | sed $'s/^/\t>> /'
echo ""

echo -e "\033[1;36m>> Running \"goimports\" check...\033[0m\n"
fail_imports=()
for file in $staged_files; do
  if [[ -f "$file" ]]; then
    output=$(goimports -l "$file")
    if [[ -n "$output" ]]; then
      fail_imports+=("$file")
    fi
  fi
done

if [[ ${#fail_imports[@]} -ne 0 ]]; then
  echo -e "❌ These files need goimports formatting:\n"
  printf '%s\n' "${fail_imports[@]}"
  echo -e "\n💡 Fix with: goimports -w <file>\n"
  print_failure
  exit 1
else
  echo -e "✅ goimports check passed\n"
fi

echo -e "\033[1;36m>> Running \"staticcheck\" per package...\033[0m\n"

# Filter out non-buildable directories
all_dirs=$(echo "$staged_files" | xargs -n1 dirname | sort -u)
pkg_dirs=()

for dir in $all_dirs; do
  if go list "./$dir" >/dev/null 2>&1; then
    pkg_dirs+=("$dir")
  else
    echo -e "⚠️  Skipping package: $dir (no buildable Go files)"
  fi
done

staticcheck_failed=0

if [[ ${#pkg_dirs[@]} -gt 0 ]]; then
  for dir in "${pkg_dirs[@]}"; do
    if ! staticcheck "./$dir"; then
      echo -e "❌ staticcheck failed in $dir\n"
      staticcheck_failed=1
    else
      echo -e "✅ staticcheck passed in $dir"
    fi
  done
else
  echo -e "⚠️  No buildable Go packages found in staged files."
fi

if [[ $staticcheck_failed -ne 0 ]]; then
  print_failure
  exit 1
fi

echo -e "\n\033[1;36m>> Running \"go test\" on impacted packages...\033[0m\n"
test_failed=0

if [[ ${#pkg_dirs[@]} -gt 0 ]]; then
for dir in "${pkg_dirs[@]}"; do
  if ! go test "./$dir"; then
    echo -e "❌ Tests failed in $dir"
    test_failed=1
  else
    echo -e "✅ Tests passed in $dir"
  fi
done
else
  echo -e "⚠️  No buildable Go packages found in staged files."
fi

if [[ $test_failed -ne 0 ]]; then
  echo -e "❌ Commit blocked due to test failures\n"
  print_failure
  exit 1
fi

print_success
exit 0