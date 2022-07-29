go           ?= go
gofmt        ?= $(go)fmt
pkgs          = ./...


help: Makefile
	@echo
	@echo " Choose a command run in Rabbit:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo


## install_revive: Install revive for linting.
.PHONY: install_revive
install_revive:
	@echo ">> ============= Install Revive ============= <<"
	$(go) install github.com/mgechev/revive@latest


## style: Check code style.
.PHONY: style
style:
	@echo ">> ============= Checking Code Style ============= <<"
	@fmtRes=$$($(gofmt) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmtRes}"; echo; \
		echo "Please ensure you are using $$($(go) version) for formatting code."; \
		exit 1; \
	fi


## check_license: Check if license header on all files.
.PHONY: check_license
check_license:
	@echo ">> ============= Checking License Header ============= <<"
	@licRes=$$(for file in $$(find . -type f -iname '*.go' ! -path './vendor/*') ; do \
			   awk 'NR<=3' $$file | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
	   done); \
	   if [ -n "$${licRes}" ]; then \
			   echo "license header checking failed:"; echo "$${licRes}"; \
			   exit 1; \
	   fi


## test_short: Run test cases with short flag.
.PHONY: test_short
test_short:
	@echo ">> ============= Running Short Tests ============= <<"
	$(go) clean -testcache
	$(go) test -mod=readonly -short $(pkgs)


## test: Run test cases.
.PHONY: test
test:
	@echo ">> ============= Running All Tests ============= <<"
	$(go) clean -testcache
	$(go) test -mod=readonly -bench=. -benchmem -v -cover $(pkgs)


## lint: Lint the code.
.PHONY: lint
lint:
	@echo ">> ============= Lint All Files ============= <<"
	revive -config config.toml -exclude vendor/... -formatter friendly ./...


## verify: Verify dependencies
.PHONY: verify
verify:
	@echo ">> ============= List Dependencies ============= <<"
	$(go) list -m all
	@echo ">> ============= Verify Dependencies ============= <<"
	$(go) mod verify


## format: Format the code.
.PHONY: format
format:
	@echo ">> ============= Formatting Code ============= <<"
	$(go) fmt $(pkgs)


## vet: Examines source code and reports suspicious constructs.
.PHONY: vet
vet:
	@echo ">> ============= Vetting Code ============= <<"
	$(go) vet $(pkgs)


## coverage: Create HTML coverage report
.PHONY: coverage
coverage:
	@echo ">> ============= Coverage ============= <<"
	rm -f coverage.html cover.out
	$(go) test -mod=readonly -coverprofile=cover.out $(pkgs)
	go tool cover -html=cover.out -o coverage.html


## run: Run the Server
.PHONY: run
run:
	@echo ">> ============= Run API Server ============= <<"
	$(go) run craft.go server -c config.dist.yml


## ci: Run all CI tests.
.PHONY: ci
ci: style check_license test vet lint
	@echo "\n==> All quality checks passed"


.PHONY: help
