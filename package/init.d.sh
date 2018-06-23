#!/bin/bash
### BEGIN INIT INFO
# Provides:          boji
# Required-Start:    $local_fs $remote_fs $network
# Required-Stop:     $local_fs $remote_fs $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start or stop boji.
# Description:       Enable service provided by boji.
### END INIT INFO

NAME=boji
DAEMON=/usr/local/bin/$NAME
USER=root
STOP_TIMEOUT=30
PIDFILE=/var/run/$NAME

export PATH="${PATH}"
[ -e /etc/default/$NAME ] && . /etc/default/$NAME

start_daemon() 
{
    # if a pidfile already exists
    if [ -f ${PIDFILE} ];
    then

        # see if the pid is still good
        pid=$(cat ${PIDFILE})
        ps -p $pid &> /dev/null
        running=$?

        # if so, ditch here. We're fine.
        if [ ${running} ];
        then
            exit 0
        fi

        # otherwise continue with starting the process.
    fi    
    
    "${DAEMON}" &> /dev/null & 
    pid=$!
    echo ${pid} > "${PIDFILE}"
    disown ${pid}
}

stop_daemon() 
{
    if [ -f ${PIDFILE} ];
    then
        kill $(cat ${PIDFILE})
        rm ${PIDFILE}
    fi
}

case "$1" in
    start)
        echo "Starting"
        start_daemon
        ;;
    stop)
        echo "Stopping"
        stop_daemon
        ;;
    reload|restart|force-reload)
        echo "Restarting" 
        stop_daemon       
        start_daemon
        ;;
    status)
        echo "undefined"
        ;;
    *)
        echo "Usage: /etc/init.d/$NAME {start|stop|restart}"
        exit 2
        ;;
esac