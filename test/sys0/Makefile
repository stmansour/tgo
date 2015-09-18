all: clean test
	@echo "*** COMPLETED in test/sys0 ***"

.PHONY:  test

clean:
	rm -f *.out *.log qmstr* x y
	@echo "*** CLEAN COMPLETE in test/sys0 ***"

test:
	./testlocal.sh
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in test/sys0 ***"
