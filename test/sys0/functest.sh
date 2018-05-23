#!/bin/bash
# This is a quick functional test simulator for tgo
# It starts up an uhura and spins through its states. Makes sure
# that it gets the responses it expects from uhura.

UPORT=8150
SCRIPTLOG="functest.log"
UVERBOSE=
UDRYRUN="-n"
ENV_DESCR="uhura_map.json"

shutdown() {
	bash ${TOOLS_DIR}/uhura_shutdown.sh -p {$UPORT} >>${SCRIPTLOG} 2>&1
	# Give the server a second to shutdown
	sleep 1
}

#---------------------------------------------------------------------
#  Find accord bin...
#---------------------------------------------------------------------
echo "Find accord/bin"
if [ -d /usr/local/accord/bin ]; then
	ACCORDBIN=/usr/local/accord/bin
	TOOLS_DIR="/usr/local/accord/testtools"
elif [ -d /c/Accord/bin ]; then
	ACCORDBIN=/c/Accord/bin
	TOOLS_DIR="/c/Accord/testtools"
else
	echo "*** ERROR: Required directory /usr/local/accord/bin or /c/Accord/bin does not exist."
	echo "           Please repair installation and try again."
	exit 2
fi

#---------------------------------------------------------------------
#  hard stance now... if uhura is running on our port, stop it first
#---------------------------------------------------------------------
echo "Stop any running instance of uhura"
COUNT=$(ps -ef | grep uhura | grep -v grep | grep ${UPORT} | wc -l)
if [ ${COUNT} -gt 0 ]; then
	echo "*** NOTICE: attempting to stop uhura already running on port ${UPORT}..."
	${TOOLS_DIR}/uhura_shutdown.sh -p ${UPORT}
	COUNT=$(ps -ef | grep uhura | grep -v grep | grep ${UPORT} | wc -l)
	if [ ${COUNT} -gt 0 ]; then
		echo "*** cannot stop it.  exiting..."
		exit 6
	fi
fi

#---------------------------------------------------------------------
#  Launch uhura and give it a second to startup. We copy the json file
#  to tmp because the path names to "this" directory are different on the
#  build machines. So, rather than try to always figure out the right path
#  name to use, we use /tmp as that just always works.
#---------------------------------------------------------------------
echo "Launch accord/bin/uhura"
if [ ! -f ${ACCORDBIN}/uhura ]; then
	echo "${ACCORDBIN}/uhura does not exist. You must create it first."
	exit 2;
fi

rm -f qm* *.log *.out
cp ${ENV_DESCR} /tmp/
echo "${ACCORDBIN}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e /tmp/${ENV_DESCR} >uhura.out 2>&1 &" >>${SCRIPTLOG} 2>&1
${ACCORDBIN}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR} >uhura.out 2>&1 &
sleep 1

../../tgo -d
shutdown

echo "BEGIN LOGFILE ANALYSIS..."
#---------------------------------------------------------------------
#  compare the output logs of previous known-good runs (.gold) to 
#  the logs produced this run (.log)
#  use the successive filter model developed in uhura/test/stateflow_normal
#  The code has been cleaned up here...
#---------------------------------------------------------------------
declare -a uhura_filters=(
	's/(20[1-4][0-9]\/[0-1][0-9]\/[0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] )(.*)/$2/'
	's/Tstamp:.*/Tstamp: TIMESTAMP/'
	's/master mode on port [0-9]+/Current working directory = SOMEdirectory/'
	's/^Current working directory = [\/a-zA-Z0-9]+/master mode on port SOMEPORT/'
	's/^exec [\/_\.a-zA-Z0-9]+ [\/_\.\-a-zA-Z0-9]+ [\/\._a-zA-Z0-9]+.*/exec SOMEPATH/g'
	's/^Uhura starting on:.*/URL: somehost:someport/'
	's/replied: \&\{OK 0.*/replied: \&\{OK <SOME_TIMESTAMP>/'
)
echo "Checking uhura logs"
cp uhura.gold x
cp uhura.log y
for f in "${uhura_filters[@]}"
do
	perl -pe "$f" x > x1; mv x1 x
	perl -pe "$f" y > y1; mv y1 y
done

UDIFFS=$(diff x y | wc -l)

#---------------------------------------------------------------------
# There is some randomness in where go routines get their timeslice.
# This results in responses coming into the logs in different sequences.
# Isolate these, and check to see if they're OK -- that is if they
# appear in uhura_variants, the list of allowable differences...
#---------------------------------------------------------------------
declare -a uhura_variants=(
	'OK <SOME_TIMESTAMP>'
	'Tgo response received'
	'Tgo @ http://localhost:8152/ replied: &{OK <SOME_TIMESTAMP>'
	'Status Handler'
)
if [ ${UDIFFS} -gt 0 ]; then
	diff x y | grep "^[<>]" | perl -pe "s/^[<>]//" | sort |uniq >z
	MISMATCHES=0
	while read p; do
		FOUND=0
		for f in "${uhura_variants[@]}"
		do
			if [[ "${p}" =~ "${f}" ]]; then
				FOUND=1
			fi
		done
		if [ ${FOUND} -eq 0 ]; then
			echo "ERROR on: ${p}"
			MISMATCHES=$((MISMATCHES+1))
		fi
	done < z
	UDIFFS=${MISMATCHES}
fi

if [ ${UDIFFS} -eq 0 ]; then
	echo "PHASE 1: PASSED"
else
	echo "PHASE 1: FAILED:  differences are as follows:"
	diff x y
	exit 1
fi

# declare -a tgo_filters=(
# 	's/(20[1-4][0-9]\/[0-1][0-9]\/[0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] )(.*)/$2/'
# 	's/Command:TESTNOW CmdCode:0 Timestamp:.*/Command:TESTNOW CmdCode:0 Timestamp: <SOME_TIMESTAMP>/'
# 	's/"UhuraURL":"http:\/\/[^\/]+\//"UhuraURL":"http:\/\/SOMEURL:SOMEPORT\//'
# )

# cp tgo.gold v
# cp tgo.log w
# for f in "${tgo_filters[@]}"
# do
# 	perl -pe "$f" v > v1; mv v1 v
# 	perl -pe "$f" w > w1; mv w1 w
# done

# UDIFFS=$(diff w v | wc -l)

# #---------------------------------------------------------------------
# # Similar randomness as what we encountered in uhura's logs
# #---------------------------------------------------------------------

# declare -a tgo_variants=(
#         'Command:TESTNOW CmdCode:0 Timestamp: <SOME_TIMESTAMP>'
#         'Comms Handler'
#         'StateUnknown: exiting 0'
#         'StateInit: exiting 0'
#         'StateReady: exiting 0'
#         'Tgo response received'
#         'Received comms from Uhura:  {Command:TESTNOW CmdCode:0 Timestamp: <SOME_TIMESTAMP>'
#         'Orchestrator: Posted READY status to uhura. ReplyCode: 0'
#         'Orchestrator: Entering StateReady'
# )
# if [ ${UDIFFS} -gt 0 ]; then
#         diff v w | grep "^[<>]" | perl -pe "s/^[<>]//" |sort | uniq >u
#         MISMATCHES=0
#         while read p; do
#                 FOUND=0
#                 for f in "${tgo_variants[@]}"
#                 do  
#                         if [[ "${p}" =~ "${f}" ]]; then
#                                 FOUND=1
#                         fi  
#                 done
#                 if [ ${FOUND} -eq 0 ]; then
#                         echo "ERROR on: ${p}"
#                         MISMATCHES=$((MISMATCHES+1))
#                 fi  
#         done < u 
#         UDIFFS=${MISMATCHES}
# fi

# if [ ${UDIFFS} -eq 0 ]; then
# 	echo "PHASE 2: PASSED"
# else
# 	echo "PHASE2 FAILED:  differences are as follows:"
# 	diff v w
# 	exit 1
# fi

echo "TGO SIMULATED FUNCTIONAL TESTS PASSED"
exit 0
