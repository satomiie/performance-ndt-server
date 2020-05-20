#!/bin/bash

### Variables ###
container_name="ndt-server"
host_ip="0.0.0.0"
check_docker_container=$(docker ps -a --no-trunc --filter name=^/${container_name}\$ | egrep -v 'CONTAINER')

clear

echo "#############################################"
echo "###      Performance Testing Service      ###"
echo "#############################################"
echo

### Start Function ###
start()
{
  echo "*************************"
  echo "*** Starting Services ***"
  echo "*************************"
  echo

  ### Running Docker Container ###
  if [[ -z "${check_docker_container}"  ]]; then
      echo "[INFO]    - Starting '${container_name}' container..."

      docker run  -d --name ${container_name} --network=host --privileged \
             --volume `pwd`/certs:/certs:ro         \
             --volume `pwd`/datadir:/var/spool/ndt  \
             --volume `pwd`/var-local:/var/local    \
             --user `id -u`:`id -g`                 \
             --cap-drop=all                         \
             ndt-server                             \
             -cert /certs/cert.pem                  \
             -key /certs/key.pem                    \
             -datadir /datadir                      \
             -ndt7_addr ${host_ip}:4443             \
             -ndt5_addr ${host_ip}:3001             \
             -ndt5_wss_addr ${host_ip}:3010

      [[ $? -eq 0  ]] || { echo "[ERROR]   - Container start failed! Please verify. Exiting..."; echo; sleep 2; exit 1; }
      echo "[SUCCESS] - Container started successfully!"
  else
      echo "[ERROR]   - Container '${container_name}' already running! Please verify. Exiting..."; echo; sleep 2; exit 1;

  fi

  ### Starting iPerf Daemon ###
  echo; echo "[INFO]    - Starting iPerf Server Services (UDP)"
  /bin/iperf -s -u -D
  [[ $? -eq 0  ]] || { echo "[ERROR]   - iPerf start failed! Please verify. Exiting..."; echo; sleep 2; exit 1; }
  echo "[SUCCESS] - iPerf started successfully!"
  sleep 2; echo; exit 0
}

### Stop Function ###
stop()
{
  echo "*************************"
  echo "*** Stopping Services ***"
  echo "*************************"
  echo

  if [[ ! -z "${check_docker_container}"  ]]; then
      echo "[INFO]    - Stopping Container '${container_name}'"
      docker stop ${container_name}
      [[ $? -eq 0  ]] || { echo "[ERROR]   - Container '${container_name}' stop failed! Please verify. Exiting..."; echo; sleep 2; exit 1; }
      echo "[SUCCESS] - '${container_name}' Stopped!"

  else
      echo "[INFO]    - Container '${container_name}' not running, skipping stop..."

  fi

  echo; echo "[INFO]    - Stopping iPerf Daemon"
  ps -ef | grep "/bin/iperf" | egrep -v grep | awk '{print "kill -9 "$2}'|sh
  check_iperf_daemon=$(ps -ef | grep '/bin/iperf' | egrep -v grep)

  [[ -z "${check_iperf_daemon}" ]] || { echo "[ERROR]   - iPerf stop failed! Please verify. Exiting..."; echo; sleep 2; exit 1; }
      echo "[SUCCESS] - iPerf Daemon Stopped!"
      sleep 2; echo; exit 0
}

### Case Statement ###
case "$1" in

        start|Start|START)
                start
        ;;

        stop|Stop|STOP)
                stop
        ;;

        *)
                echo "Usage: $0 (start|stop)"
        ;;
esac
