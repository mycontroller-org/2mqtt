#!/bin/sh

# user command
# command 'start' starts the service
# command 'stop' stops the service
USER_COMMAND=${1}

if [ -z "${USER_COMMAND}" ];then
  echo "there is no command specified, supported commands are [start, stop]"
  exit 1
fi

# Get user current location
USER_LOCATION=${PWD}
ACTUAL_LOCATION=`dirname $0`

# change the location to where exactly script is located
cd ${ACTUAL_LOCATION}

BINARY_FILE=./2mqtt
CONFIG_FILE=./cofig.yaml

START_COMMAND="${BINARY_FILE} -config ${CONFIG_FILE}"

SVC_PID=`ps -ef | grep "${START_COMMAND}" | grep -v grep | awk '{ print $2 }'`

if [ ${USER_COMMAND} = "start" ]; then
  if [ ! -z "$SVC_PID" ];then
    echo "there is a running instance of the 2mqtt server on the pid: ${SVC_PID}"
  else
    mkdir -p logs
    exec $START_COMMAND >> logs/2mqtt.log 2>&1 &
    echo "start command issued to the 2mqtt server"
  fi
elif [ ${USER_COMMAND} = "stop" ]; then
  if [ ${SVC_PID} ]; then
    kill -15 ${SVC_PID}
    echo "stop command issued to the 2mqtt server"
  else
    echo "2mqtt server is not running"
  fi
else
  echo "invalid command [${USER_COMMAND}], supported commands are [start, stop]"
fi

# back to user location
cd ${USER_LOCATION}