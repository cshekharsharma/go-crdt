# Makefile for Go-CRDT project

# This Makefile defines several commands to manage and build the project omega.
# The commands include linting, testing, cleaning, and running the project,
# as well as configuring the project and generating test coverage reports.

# Colors
RED         := \033[0;31m
GREEN       := \033[0;32m
YELLOW      := \033[1;33m
BLUE        := \033[0;34m
NC          := \033[0m

# Go Commands
MAKE        := make # Command to run Makefile
GOCMD       := go   # Command to run Go
GOBUILD     := $(GOCMD) build   # Command to build Go binaries
GOCLEAN     := $(GOCMD) clean   # Command to clean Go binaries and caches
GOIMPORT    := goimports        # Command to adjust/fix Go imports
GOFMT       := gofmt            # Command to format Go source code
GOINSTALL   := $(GOCMD) install # Command to install go packages
STATICCHECK := staticcheck      # Command to run static analysis on Go code

# Target: help
# Description: List all make targets with descriptions
.PHONY: help
help:
	@printf "\n${YELLOW}Makefile Targets:${NC}\n\n"
	@printf "  ${GREEN}clean${NC}          - Clean the previous builds\n"
	@printf "  ${GREEN}configure${NC}      - Configure the project (tidy, format, static-check)\n"
	@printf "  ${GREEN}install${NC}        - Alias for configure command\n"
	@printf "  ${GREEN}lint${NC}           - Run linting & static checks for go code\n"
	@printf "  ${GREEN}test${NC}           - Run unit tests\n"
	@printf "  ${GREEN}testcoverage${NC}   - Run unit tests with coverage report\n"
	@printf "  ${GREEN}all${NC}            - Run all important steps from the list\n\n"

# Target: configure
# Description: Configure the project by tidying and verifying the modules,
# formatting the code, and running static analysis.
.PHONY: configure
configure:
	@printf "\n${YELLOW}CONFIGURING THE PACKAGE...${NC}\n\n"
	export GOPRIVATE=bitbucket.shl.systems
	$(GOCMD) mod tidy
	$(GOCMD) mod verify
	$(GOIMPORT) -w .
	$(GOFMT) -s -w .
	$(STATICCHECK) ./...
	cp ./tools/pre-commit.sh ./.git/hooks/pre-commit
	chmod +x ./.git/hooks/pre-commit
	cp ./tools/pre-push.sh ./.git/hooks/pre-push
	chmod +x ./.git/hooks/pre-push
	@printf "\n✅ Project successfully configured.\n"

# Target: install
# Description: Install the dependencies and configure the project (alias for configure),
# formatting the code, and running static analysis.
.PHONY: install
install: configure

# Target: lint
# Description: Run static code analysis using golangci-lint
.PHONY: lint
lint:
	@printf "\n${YELLOW}LINTING THE CODEBASE...${NC}\n\n"
	$(GOIMPORT) -w .
	$(GOFMT) -s -w .
	$(STATICCHECK) ./...
	@printf "\n✅ Project successfully linted and analysed.\n"

# Target: test
# Description: Run unit tests for the project.
.PHONY: test
test:
	@printf "\n${YELLOW}RUNNING UNIT TESTS...${NC}\n\n"
	rm -rf ./coverage.txt
	$(GOCMD) run tools/gotest_exec.go --skip-mocks
	@printf "\n✅ Testcase execution completed.\n"

# Target: testcoverage
# Description: Run unit tests and generate a coverage report for the project.
.PHONY: testcoverage
testcoverage:
	@printf "\n${YELLOW}RUNNING UNIT TESTS WITH COVERAGE REPORT...${NC}\n\n"
	rm -rf ./coverage.txt
	$(GOCMD) run tools/gotest_exec.go --skip-mocks
	$(GOCMD) run tools/gotest_coverage.go
	@printf "\n✅ Testcase execution completed with coverage report.\n\n"

# Target: clean
# Description: Clean the previous builds and remove the binary.
.PHONY: clean
clean:
	@printf "\n${YELLOW}CLEANING THE ENVIRONMENT...${NC}\n\n"
	$(GOCLEAN)
	@printf "\n✅ Project cleaned.\n"


# Target: all
# Description: configure and test the package
.PHONY: all
all:
	@$(MAKE) clean
	@$(MAKE) configure
	@$(MAKE) testcoverage
	@printf "\n✅ All steps completed successfully.\n\n"
