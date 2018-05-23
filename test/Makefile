TOP=..
THISDIR=test
TESTREPORT="testreport.txt"
DIRS=simfunc func1 sys0

all: clean test
	@echo "*** COMPLETED in ${THISDIR} ***"

.PHONY:  test

clean:
	rm -rf echosrv echosrv_test
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** CLEAN COMPLETE in ${THISDIR} ***"

test:
	for dir in $(DIRS); do make -C $$dir test;done
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in ${THISDIR} ***"
