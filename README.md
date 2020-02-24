# CADCloud
CAD/EDA server

![Image of landing](https://raw.githubusercontent.com/CADCloud/CADCloud/master/screenshot/landing.png)

This project aims to create a version tracking and collaboration tool to Open Hardware communities. It does "support" FreeCAD 0.19 (current developer version) with the CLOUD workbench pre-compiled. To use it you need to either import a step file into FreeCAD and export it to the server through the following command into the python console 

import Cloud \
Cloud.cloudurl(u"https://YOUR SERVER URI") \
Cloud.cloudtcpport(u"443") \
Cloud.cloudaccesskey(u"YOUR ACCESS KEY") \
Cloud.cloudsecretkey(u"YOUR PRIVATE KEY") \
Cloud.cloudsave(u"YOUR MODEL NAME (lowercase only)") \

You can export native FreeCAD file. The WebGL rendering is alpha stage, but still could work ;). 
In any cases your model will be saved on the server side through the amazon S3 protocol  

To read back a model  

Cloud.cloudrestore(u"YOUR MODEL NAME (lowercase only)") 

# Docker build

The Docker container requires at least 2 CPUs, 4 GB, 4 GB of RAM / Storage. This is coming from the fact that
it is going to run FreeCAD and minio locally

Please edit the start_container file and add the following data to your relevant server

export SMTP_SERVER=
export SMTP_PASSWORD=
export SMTP_ACCOUNT=

CADCloud requires email validation and needs to be able to send email for testing

To build the container initial self-signed certificate must be generated the process is describe into build_docker.
If you are on linux or MacOS with openssl tools installed you can use that script straight forward.

The build process is made in 2 steps due to the fact that CADCloud embedded a specific developper version of FreeCAD 0.19
snap released on the Ubuntu snap store as test/beta. That version has the Cloud workbench activated.

After the build_docker script has been executed start the container with the following command

 docker run --privileged --name cadcloud  -p 443:443 cadcloud

The privileged mode is required to get snap working properly

Then connect to the container through

docker exec -it cadcloud /bin/bash

issue the ./start_container command. You can make it work in background if you want. That command is going to download
the snap, install it and make it available to the system. It will also install the latest minio build available. When the
execution is done it will kick the various CADCloud daemons.

You shall now be able to enjoy a local instance of CADCloud. Got to Chrome and issue https://127.0.0.1

Enjoy, debug and issue PR !


