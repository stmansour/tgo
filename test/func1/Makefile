all: clean test
	@echo "*** COMPLETED in test/func1 ***"


clean:
	go clean
	rm -f *.out *.log qmstr* [u-z]
	@echo "*** CLEAN COMPLETE in test/func1 ***"

test:	clean
	M=`ps -ef | grep echosrv | grep -v grep | wc -l`;if [ $M -gt 0 ]; then pkill echosrv; fi
	cp -r ../../../echosrv/package/echosrv ../../../echosrv/package/echosrv_test ../ 
	./func.sh
	@echo "*** TEST COMPLETE - ALL TESTS PASSED in test/func1 ***"

pretty:
	cat uhura_map.json | python -m json.tool > u; mv u uhura_map.json
