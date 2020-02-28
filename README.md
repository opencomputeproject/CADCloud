# CADCloud
CAD/EDA server

![Image of landing](https://raw.githubusercontent.com/CADCloud/CADCloud/master/screenshot/landing.png)

## Description
This project aims to create a version tracking and collaboration tool to Open Hardware communities. 

## Prerequisites
It does "support" FreeCAD 0.19 (current developer version) with the CLOUD workbench pre-compiled. 

## Usage
To use it you need to either import a STEP file into FreeCAD and export it to the server through the following command in the python console:

```python
import Cloud \
Cloud.cloudurl(u"https://YOUR SERVER URI") \
Cloud.cloudtcpport(u"443") \
Cloud.cloudaccesskey(u"YOUR ACCESS KEY") \
Cloud.cloudsecretkey(u"YOUR PRIVATE KEY") \
Cloud.cloudsave(u"YOUR MODEL NAME (lowercase only)") \
```

Note: You can export a native FreeCAD file. Though be aware that WebGL rendering is still alpha, though it still could work ;).  
In any cases your model will be saved on the server side through the amazon S3 protocol  

To read back a model  

```python
Cloud.cloudrestore(u"YOUR MODEL NAME (lowercase only)") 
```

## Docker build

The Docker container requires at least 2 CPUs, 4 GB, 4 GB of RAM / Storage. This is due to the fact that it is runs FreeCAD and minio locally.

Please edit the `start_container` file and add the following data to your relevant server

```bash
export SMTP_SERVER=
export SMTP_PASSWORD=
export SMTP_ACCOUNT=
```

Note: CADCloud requires email validation and needs to be able to send email for testing

To build the container initial self-signed certificate must be generated the process is described in `build_docker`.
If you are on linux or MacOS with openssl tools installed you can use that script straight forward.

The build process is made in 2 steps due to the fact that CADCloud embedded a specific developer version of FreeCAD 0.19
snap released on the Ubuntu snap store as test/beta. That version has the Cloud workbench activated.

After the `build_docker` script has been executed start the container with the following command

```bash
docker run --privileged --name cadcloud  -p 443:443 cadcloud
```

The `-privileged` mode is required to get snap working properly

Then connect to the container through

```bash
docker exec -it cadcloud /bin/bash
```

Invoke the `./start_container` command.  
Note: it may also  be invoked in background like so: `./start_container&` 

`./start_container` will download the snap, install it and make it available to the system. It will also install the latest minio build available. When the
execution is done it will spawn the various CADCloud daemons.

You now will be able to enjoy a local instance of CADCloud! 

In your preferred web browser type: `https://127.0.0.1` in the URL field. 

Enjoy, debug and issue PR(s) !

