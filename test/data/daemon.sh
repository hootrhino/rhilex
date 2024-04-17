#!/bin/sh
# Create Time: 2023-11-27 14:59:06

WORKING_DIRECTORY="/usr/local"
PID_FILE="/var/run/rhilex.pid"
EXECUTABLE_PATH="$WORKING_DIRECTORY/rhilex"
CONFIG_PATH="$WORKING_DIRECTORY/rhilex.ini"

log() {
    local level=$1
    shift
    echo "[$level] $(date +'%Y-%m-%d %H:%M:%S') - $@"
}

start() {
    rm -f /var/run/rhilex-stop.sinal
    pid=$(pgrep -x -n -f "/usr/local/rhilex run -config=/usr/local/rhilex.ini")
    if [ -n "$pid" ]; then
        log INFO "rhilex is running with Pid:${pid}"
        exit 0
    fi
    daemon &
    exit 0
}

stop() {
    echo "1" > /var/run/rhilex-stop.sinal
    if pgrep -x "rhilex" > /dev/null; then
        log INFO "rhilex process is running. Killing it..."
        pkill -x "rhilex"
        log INFO "rhilex process has been killed."
    else
        log WARNING "rhilex process is not running."
    fi
}

restart() {
    stop
    sleep 1
    start
}

status() {
    log INFO "Checking rhilex status."
    pid=$(pgrep -x -n "rhilex")
    if [ -n "$pid" ]; then
        log INFO "rhilex is running with Pid:${pid}"
    else
        log INFO "rhilex is not running."
    fi
}

daemon() {
    while true; do
        if pgrep -x "rhilex" > /dev/null; then
            sleep 3
            continue
        fi
        if ! pgrep -x "rhilex" > /dev/null; then
            if [ -e "/var/run/rhilex-upgrade.lock" ]; then
                log INFO "File /var/run/rhilex-upgrade.lock exists. May upgrade now."
                sleep 2
                continue
            elif [ -e "/var/run/rhilex-stop.sinal" ]; then
                log INFO "/var/run/rhilex-stop.sinal file found. Exiting."
                exit 0
            else
                log WARNING "Detected that rhilex process is interrupted. Restarting..."
                /usr/local/rhilex run -config=/usr/local/rhilex.ini
                log WARNING "Detected that rhilex process has Restarted."
            fi
        fi
        sleep 4
    done
}

case "$1" in
    start)
        start
    ;;
    restart)
        restart
    ;;
    stop)
        stop
    ;;
    status)
        status
    ;;
    *)
        log ERROR "Usage: $0 {start|restart|stop|status}"
        exit 1
    ;;
esac

