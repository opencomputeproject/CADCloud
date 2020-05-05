#!/bin/bash
# snapcraft yaml automatic generation tool for FreeCAD 0.19
# (c) 2019 vejmarie 
# Distributed under LGPL v2.0
# Please report bugs at vejmarie@ruggedpod.qyshare.com

rm snapcraft.yaml
lib_list=`ls /opt/local/FreeCAD-0.19/lib/*.so`
mod_list=`ls -ltd $(find /opt/local/FreeCAD-0.19/Mod) | awk '{ print $9}'`
ext_list=`ls -ltd $(find /opt/local/FreeCAD-0.19/Ext) | awk '{ print $9}'`
bin_list=`ls -ltd $(find /opt/local/FreeCAD-0.19/bin) | awk '{ print $9}'`
data_list=`ls -ltd $(find /opt/local/FreeCAD-0.19/data) | awk '{ print $9}'`
share_list=`ls -ltd $(find /opt/local/FreeCAD-0.19/share/Mod) | awk '{ print $9}'`
python_pivy=`ls -ltd $(find /usr/local/lib/python3.5) | awk '{ print $9}'`
python_list=`ls -ltd $(find /usr/lib/python3.5) | awk '{ print $9}'`
python_list_2=`ls -ltd $(find /usr/lib/python3) | awk '{ print $9}'`
pyside_list=`ls -ltd $(find  /usr/share/PySide/typesystems) | awk '{ print $9}'`
pyside2_list=`ls -ltd $(find  /usr/share/PySide2/typesystems) | awk '{ print $9}'`
python_binary=`ls /usr/bin/python3 /usr/bin/python3.5 /usr/bin/py*3.5`
syslib_qt=`ls -ltd $(find /usr/lib/x86_64-linux-gnu/qt5) | awk '{ print $9}'`
syslib=`ls -ltd $(find /usr/lib/x86_64-linux-gnu/dri) | awk '{ print $9}'`
syslib2=`ls -ltd $(find  /usr/lib/x86_64-linux-gnu/gtk-2.0) | awk '{ print $9}'`
syslib3=`ls -ltd $(find  /usr/lib/x86_64-linux-gnu/gdk-pixbuf-2.0) | awk '{ print $9}'`
syslib4=`ls -ltd $(find  /usr/lib/locale) | awk '{ print $9}'`
syslib5=`ls -ltd $(find  /usr/share/X11/locale) | awk '{ print $9}'`
syslib6=`ls -ltd $(find  /usr/lib/x86_64-linux-gnu/*gtk*) | awk '{ print $9}'`
syslib7=`ls -ltd $(find  /usr/share/glib-2.0) | awk '{ print $9}'`
syslib8=`ls -ltd $(find  /usr/share/i18n/) | awk '{ print $9}'`
syslib9=`ls /usr/bin/locale*`
syslib10=`ls -ltd $(find  /usr/share/matplotlib/mpl-data/) | awk '{ print $9}'`
syslib11=`ls -ltd $(find  /usr/lib/x86_64-linux-gnu/gio) | awk '{ print $9}'`
bin_list2=`ls /usr/bin/dconf`
dconf_list=`ls  -ltd $(find  /usr/lib/dconf) | awk '{ print $9}'`
fonts_list=`ls  -ltd $(find  /usr/share/fonts/truetype/dejavu) | awk '{ print $9}'`
calculix_list=`ls /usr/bin/ccx /usr/lib/x86_64-linux-gnu/libspool* /usr/lib/libmp* /usr/lib/libarpack* /usr/lib/liblapack* /usr/lib/x86_64-linux-gnu/libgfort* /usr/lib/x86_64-linux-gnu/libgom* /usr/lib/libblas.so* /usr/lib/x86_64-linux-gnu/libquadm* /usr/lib/libopen-* /usr/lib/x86_64-linux-gnu/libhw* /usr/lib/x86_64-linux-gnu/libnuma* /usr/lib/x86_64-linux-gnu/libltdl*`
mpi_list=`ls  -ltd $(find  /usr/lib/openmpi*) | awk '{ print $9}'`
# icon_list=`ls  -ltd $(find  /usr/share/icons/[a-j]*) | awk '{ print $9}'`
# icon_list2=`ls  -ltd $(find  /usr/share/icons/[k-z]*) | awk '{ print $9}'`
# mime_list=`ls  -ltd $(find  /usr/share/mime) | awk '{ print $9}'`
# theme_list=`ls  -ltd $(find  /usr/share/themes) | awk '{ print $9}'`
echo "
name: freecad
version: 0.19
summary: development version 
description: development version This version is far from being bug free but integrates the latest FreeCAD technologies
parts:
 example-part:
    plugin: copy
    files:
      bin/launcher : bin/launcher
      fontconfig/fonts.conf : fontconfig/fonts.conf
      etc/matplotlibrc : etc/matplotlibrc" >> snapcraft.yaml
#      bin/FreeCAD : opt/local/FreeCAD-0.19/bin/FreeCAD" >> snapcraft.yaml
for i in $bin_list $bin_list2 $dconf_list /usr/bin/fc-cache $mpi_list
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
for i in $lib_list $calculix_list
do
        second_name=`echo $i | sed 's/\///'`
	echo "      $i : $second_name" >> snapcraft.yaml
done
for i in $mod_list $ext_list $theme_list
do
        if [ ! -d "$i" ]
	then
		second_name=`echo $i | sed 's/\///'`
       		echo "      $i : $second_name" >> snapcraft.yaml
	fi
done
for i in $data_list $share_list
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
for i in $python_list $python_list_2 $python_pivy $pyside_list $pyside2_list
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
for i in $python_binary $icon_list2
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
for i in  $syslib $syslib2 $syslib3 $syslib4 $syslib5 $syslib6 $syslib7 $syslib8 $syslib9 $syslib10 $syslib11 $fonts_list $icon_list $mime_list
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
for i in  $syslib_qt
do
        if [ ! -d "$i" ]
        then
                second_name=`echo $i | sed 's/\///'`
                echo "      $i : $second_name" >> snapcraft.yaml
        fi
done
# python:
#    plugin: python2
echo "
    snap:
     - bin/launcher
     - fontconfig/fonts.conf
     - etc/matplotlibrc
     - opt/local/FreeCAD-0.19/bin/FreeCAD
     - opt
     - usr" >> snapcraft.yaml
for i in $lib_list
do
        second_name=`echo $i | sed 's/\///'`
        echo "     - $second_name" >> snapcraft.yaml
done
echo "
apps:
 freecad:
  command: bin/launcher
  plugs: [ locale-control,x11,opengl,network-bind,home,unity7 ]
" >> snapcraft.yaml
