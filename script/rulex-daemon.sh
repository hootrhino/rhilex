#!/bin/bash

SERVICE_NAME="rhilex"
WORKING_DIRECTORY="/usr/local"
EXECUTABLE_PATH="$WORKING_DIRECTORY/$SERVICE_NAME"
CONFIG_PATH="$WORKING_DIRECTORY/$SERVICE_NAME.ini"

SERVICE_FILE="/etc/init.d/$SERVICE_NAME.service"

STOP_SIGNAL="/var/run/rhilex-stop.sinal"
UPGRADE_SIGNAL="/var/run/rhilex-upgrade.lock"

SOURCE_DIR="$PWD"


log() {
    local level=$1
    shift
    echo "[$level] $(date +'%Y-%m-%d %H:%M:%S') - $@"
}

install(){
cat > "$SERVICE_FILE" << EOL
#!/bin/sh

export PATH=\$PATH:/usr/local/bin:/usr/bin:/usr/sbin:/usr/local/sbin:/sbin

### BEGIN INIT INFO
# Provides:          rhilex
# Required-Start:    \$network \$local_fs \$remote_fs
# Required-Stop:     \$network \$local_fs \$remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: rhilex Service
# Description:       rhilex Service
### END INIT INFO

#
# Create Time: $(date +'%Y-%m-%d %H:%M:%S')
#
EXECUTABLE_PATH="$WORKING_DIRECTORY/rhilex"
CONFIG_PATH="$WORKING_DIRECTORY/rhilex.ini"

log() {
    local level=\$1
    shift
    echo "[\$level] \$(date +'%Y-%m-%d %H:%M:%S') - \$@"
}

start() {
    rm -f $UPGRADE_SIGNAL
    rm -f $STOP_SIGNAL
    pid=\$(pgrep -x -n -f "$EXECUTABLE_PATH run -config=$CONFIG_PATH")
    if [ -n "\$pid" ]; then
        log INFO "rhilex is running with Pid:\${pid}"
        exit 0
    fi
    cd $WORKING_DIRECTORY
    daemon
}

stop() {
    echo "1" > $STOP_SIGNAL
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
    pid=\$(pgrep -x -n "rhilex")
    if [ -n "\$pid" ]; then
        log INFO "rhilex is running with Pid:\${pid}"
    else
        log INFO "rhilex is not running."
    fi
}

daemon() {
    while true; do
        if pgrep -x "rhilex" > /dev/null; then
            log INFO "rhilex process exists"
            sleep 3
            continue
        fi
        if ! pgrep -x "rhilex" > /dev/null; then
            if [ -e "$UPGRADE_SIGNAL" ]; then
                log INFO "File $UPGRADE_SIGNAL exists. May upgrade now."
                sleep 2
                continue
            elif [ -e "$STOP_SIGNAL" ]; then
                log INFO "$STOP_SIGNAL file found. Exiting."
                exit 0
            else
                log WARNING "Detected that rhilex process is interrupted. Restarting..."
                cd $WORKING_DIRECTORY
                $EXECUTABLE_PATH run -config=$CONFIG_PATH
                log WARNING "Detected that rhilex process has Restarted."
            fi
        fi
        sleep 4
    done
}

case "\$1" in
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
        log ERROR "Usage: \$0 {start|restart|stop|status}"
        exit 1
    ;;
esac

EOL

    mkdir -p $WORKING_DIRECTORY
    chmod +x $SOURCE_DIR/rhilex
    log INFO "Copy rhilex to $WORKING_DIRECTORY"
    cp -rfp "$SOURCE_DIR/rhilex" "$EXECUTABLE_PATH"

    log INFO "Copy rhilex.ini to $WORKING_DIRECTORY"
    cp -rfp "$SOURCE_DIR/rhilex.ini" "$CONFIG_PATH"

    log INFO "Copy license.lic to $WORKING_DIRECTORY"
    cp -rfp "$SOURCE_DIR/license.lic" "$WORKING_DIRECTORY/"

    log INFO "Copy license.key to $WORKING_DIRECTORY"
    cp -rfp "$SOURCE_DIR/license.key" "$WORKING_DIRECTORY/"
    chmod 777 $SERVICE_FILE
    if [ $? -eq 0 ]; then
        log INFO "rhilex service has been created and extracted."
    else
        log ERROR "Failed to create the rhilex service or extract files."
    fi
    exit 0
}

__remove_files() {
    local file=$1
    log INFO "Removing $file."
    if [ -e "$file" ]; then
        if [ -d "$file" ]; then
            rm -rf "$file"
        else
            rm "$file"
        fi
        log INFO "$file removed."
    else
        log INFO "$file not found. No need to remove."
    fi
}

uninstall(){
    if [ -e "$SERVICE_FILE" ]; then
        $SERVICE_FILE stop
    fi
    __remove_files "$SERVICE_FILE"
    __remove_files "$WORKING_DIRECTORY/rhilex"
    __remove_files "$WORKING_DIRECTORY/rhilex.ini"
    __remove_files "$WORKING_DIRECTORY/rhilex.db"
    __remove_files "$WORKING_DIRECTORY/license.lic"
    __remove_files "$WORKING_DIRECTORY/license.key"
    __remove_files "$WORKING_DIRECTORY/rhilex_internal_datacenter.db"
    __remove_files "$WORKING_DIRECTORY/upload/"
    __remove_files "$WORKING_DIRECTORY/rhilexlog.txt"
    __remove_files "$WORKING_DIRECTORY/rhilex-daemon-log.txt"
    __remove_files "$WORKING_DIRECTORY/rhilex-recover-log.txt"
    __remove_files "$WORKING_DIRECTORY/rhilex-upgrade-log.txt"
    __remove_files "$WORKING_DIRECTORY/rhilexlog-*.txt.gz"
    log INFO "rhilex has been uninstalled."
}

start() {
    $SERVICE_FILE start
}

restart() {
    $SERVICE_FILE restart
}

stop() {
    $SERVICE_FILE stop
}

status() {
    $SERVICE_FILE status
}

case "$1" in
    install)
        install
    ;;
    start)
        start
    ;;
    restart)
        stop
        start
    ;;
    stop)
        stop
    ;;
    uninstall)
        uninstall
    ;;
    status)
        status
    ;;
    *)
        log ERROR "Usage: $0 {install|start|restart|stop|uninstall|status}"
        exit 1
    ;;
esac

exit 0
