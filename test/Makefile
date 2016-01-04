all: clean test
	@echo "*** COMPLETED in test/ ***"

.PHONY:  test

clean:
	rm -rf echosrv echosrv_test
	cd simfunc;make clean
	cd func1;make clean
	cd sys0;make clean
	@echo "*** CLEAN COMPLETE in test/ ***"

test:
	echo "\n**** SIMFUNC ****\n";cd simfunc;make test
	echo "\n**** FUNC1 ****\n";cd func1;make test
	echo "\n**** SYS0  ****\n";cd sys0;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in test/ ***"

