#!/bin/bash

X=cible
cp $X.service /lib/systemd/system/
chmod 0664 /lib/systemd/system/$X.service
systemctl daemon-reload
systemctl enable $X.service
systemctl start $X
systemctl status $X
