all: clean test
	@echo "*** COMPLETED in test/sys0 ***"

.PHONY:  test

clean:
	go clean
	rm -f *.out *.log qmstr*
	@echo "*** CLEAN COMPLETE in test/sys0 ***"

test:
	./testlocal.sh
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in test/sys0 ***"
