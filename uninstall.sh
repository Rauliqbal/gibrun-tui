#!/usr/bin/env bash

sudo rm -f /usr/local/bin/gibrun
sudo rm -f /etc/polkit-1/rules.d/49-gibrun.rules
sudo groupdel gibrun

echo "🗑️ gibrun removed"
