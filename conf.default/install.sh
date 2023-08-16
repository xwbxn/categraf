#! /bin/bash

echo 'install service'
if command -v systemctl >/dev/null 2>&1; then 
    cp ./conf/categraf.service /usr/lib/systemd/system
    systemctl enable categraf
    systemctl start categraf
    echo install complete, status: $(systemctl is-active categraf) 
else 
    cp ./conf/categraf.sh /etc/init.d/categraf
    cmhod a+x /etc/init.d/categraf
    chkconfig --add categraf
    /etc/init.d/categraf start
    echo install complete, status: $?
fi
