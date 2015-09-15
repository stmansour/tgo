all: clean tgo test
	@echo "*** COMPLETED ***"

.PHONY:  test

install: tgo
	cp uhura /usr/local/accord/bin
	@echo "*** INSTALL COMPLETED ***"

tgo: *.go
	go fmt
	go vet
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	go clean
	rm -f *.json *.out *.log qmstr* phonehome
	@echo "*** CLEAN COMPLETE ***"

test:
	@echo "http://localhost:8123/" >phonehome
	go test
	@cp test/simfunc/env1.json /tmp/;/usr/local/accord/bin/uhura -p 8123 -e /tmp/env1.json >uhura.out 2>&1 &
	sleep 1
	./tgo -d -F
	@RET=`/usr/local/accord/testtools/uhura_shutdown.sh -p 8123`;echo "uhura normal shutdown"
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"

coverage:
	go test -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out

package: clean tgo
	rm -rf ./tmp/tgo ./tmp/tgo.tar*
	mkdir -p ./tmp/tgo
	cp tgo ./tmp/tgo
	cp activate.* ./tmp/tgo
	cd ./tmp/tgo;/usr/local/accord/testtools/makephonehome.sh
	cd ./tmp;tar cvf tgo.tar tgo;gzip tgo.tar

publish: package
	cd ./tmp;deployfile.sh tgo.tar.gz jenkins-snapshot/tgo/latest

