#!/bin/bash

set -euo pipefail

print_start() {
  echo -e "\n\033[1;33m=============================="
  echo -e "  🚧 RUNNING PRE-PUSH HOOK  "
  echo -e "==============================\033[0m\n"
}

print_success() {
  echo -e "\n\033[1;32m=============================="
  echo -e "  ✅ Pre-Push hook passed."
  echo -e "==============================\033[0m\n"
}

print_failure() {
  echo -e "\n\033[1;31m=============================="
  echo -e "  ❌ Pre-Push hook failed."
  echo -e "==============================\033[0m\n"
}

trap 'echo -e "\n💥 An unexpected error occurred. Aborting push."; print_failure; exit 1' ERR

print_start

echo -e "\033[1;36m>> Checking impacted Go files for push...\033[0m\n"

impacted_files=$(git diff --name-only @{u}...HEAD | grep '\.go$' || true)

if [[ -z "$impacted_files" ]]; then
  echo -e "✅ No Go files impacted for push. Skipping checks.\n"
  print_success
  exit 0
fi

echo -e "Impacted Go files:"
echo "$impacted_files" | sed $'s/^/\t>> /'
echo ""

echo -e "\033[1;36m>> Running \"goimports\" check...\033[0m\n"
fail_imports=()
for file in $impacted_files; do
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
all_dirs=$(echo "$impacted_files" | xargs -n1 dirname | sort -u)
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
  echo -e "⚠️  No buildable Go packages found in impacted files."
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
  echo -e "⚠️  No buildable Go packages found in impacted files."
fi

if [[ $test_failed -ne 0 ]]; then
  echo -e "❌ Push blocked due to test failures\n"
  print_failure
  exit 1
fi

print_success
exit 0