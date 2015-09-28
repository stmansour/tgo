#!/bin/bash
# activation script for tgo

usage() {
    cat << ZZEOF

Usage:   activate.sh [OPTIONS] cmd
cmd is one of: START | STOP | READY 

OPTIONS
-p port number on which to listen

Examples:
Command to start application on port 8081:

	bash$  activate.sh -p 8081 START 

Command to stop the application on port 8081:

	bash$  activate.sh -p 8081 STOP

ZZEOF

	exit 0
}

start() {
	./tgo -d
}

while getopts ":p:ih:" o; do
    case "${o}" in
        h)
            usage
            ;;
        p)
            PORT=${OPTARG}
	    echo "PORT set to: ${PORT}"
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

for arg do
	# echo '--> '"\`$arg'"
    case "$arg" in
	"START")
		echo "START tgo"
		./tgo -d &
        echo "OK"
		;;
	"STOP")
		echo "STOP"
		;;
	"READY")
		echo "READY"
		;;
	*)
		echo "Unrecognized command: $arg"
		exit 1
		;;
    esac
done
