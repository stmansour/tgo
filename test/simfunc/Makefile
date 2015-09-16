all: clean test
	@echo "*** COMPLETED in test/simfunc ***"

.PHONY:  test

clean:
	go clean
	rm -f *.out *.log qmstr* phonehome
	@echo "*** CLEAN COMPLETE in test/simfunc ***"

test:
	./simfunc.sh
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in test/simfunc ***"
