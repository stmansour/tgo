#!/bin/bash
# This is a quick functional test simulator for tgo
# It starts up an uhura and spins through its states. Makes sure
# that it gets the responses it expects from uhura.

UPORT=8100
SCRIPTLOG="testlocal.log"
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
rm -f qm* *.log *.out
cp ${ENV_DESCR} /tmp/
echo "${ACCORDBIN}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e /tmp/${ENV_DESCR} >uhura.out 2>&1 &" >>${SCRIPTLOG} 2>&1
${ACCORDBIN}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR} >uhura.out 2>&1 &
sleep 1

../../tgo -d -D

shutdown
exit 0
