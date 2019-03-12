#!/usr/bin/env bash

# Prevent accidental stomp of ssh files
if [ -f ~/.ssh/id_rsa ]
then
     echo "~/.ssh/id_rsa file exists.. Exiting."
     exit 1
fi

# Generate a new key each time the image is ran.
ssh-keygen -t rsa -f ~/.ssh/id_rsa -q -P ""

exec /cluster-manager

