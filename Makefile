all: clean tgo test
	@echo "*** COMPLETED ***"

.PHONY:  test

install: tgo
	@echo "*** INSTALL COMPLETED ***"

tgo: *.go
	go fmt
	go vet
	gl=$(which golint);if [ "x${gl}" != "x" ]; then golint; fi
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	go clean
	rm -f *.json *.out *.log qmstr* phonehome
	cd test;make clean
	@echo "*** CLEAN COMPLETE ***"

test:
	go test
	cd test;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"

coverage:
	go test -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out

package: tgo
	rm -rf ./tmp/tgo ./tmp/tgo.tar*
	mkdir -p ./tmp/tgo
	cp tgo ./tmp/tgo
	cp activate.* ./tmp/tgo
	cd ./tmp;tar cvf tgo.tar tgo;gzip tgo.tar

publish: package
	cd ./tmp;deployfile.sh tgo.tar.gz jenkins-snapshot/tgo/latest

