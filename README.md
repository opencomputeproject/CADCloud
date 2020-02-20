# CADCloud
CAD/EDA server

![Image of landing](https://raw.githubusercontent.com/CADCloud/CADCloud/master/screenshot/landing.png)

This project aims to create a version tracking and collaboration tool to Open Hardware communities. It does "support" FreeCAD 0.19 (current developer version) with the CLOUD workbench pre-compiled. To use it you need to either import a step file into FreeCAD and export it to the server through the following command into the python console \

import Cloud \
Cloud.cloudurl(u"https://justyour.parts") \
Cloud.cloudtcpport(u"443") \
Cloud.cloudaccesskey(u"YOUR ACCESS KEY") \
Cloud.cloudsecretkey(u"YOUR PRIVATE KEY") \
Cloud.cloudsave(u"YOUR MODEL NAME (lowercase only)") \

You can export native FreeCAD file. The WebGL rendering is alpha stage, but still could work ;). 
In any cases your model will be saved on the server side through the amazon S3 protocol  \

To read back a model \ 

Cloud.cloudrestore(u"YOUR MODEL NAME (lowercase only)") \

