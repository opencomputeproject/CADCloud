#!/bin/bash 

# Vagrant provisioning script to build up FreeCAD based on OCCT 7 and Salome 7.7.1 on Linux Ubuntu
# (c) 2016 Jean-Marie Verdun / vejmarie (vejmarie@ruggedpod.qyshare.com)
# Released under GPL v2.0
# Provided without any warranty
# Warning: compilation time is long quite long
# Must add the autlogin script
# Must add a lightdm start / reboot ?
export LC_ALL=en_US.UTF-8
FREECAD_GIT="https://github.com/FreeCAD/FreeCAD.git"
FREECAD_BRANCH="master"
export CCACHE_DISABLE=1
CPU=2
INIT_DISTRO=1

function create_deb {
rm -rf /tmp/$1-$2
mkdir /tmp/$1-$2
mkdir /tmp/$1-$2/DEBIAN
echo "Package:$1" >> /tmp/$1-$2/DEBIAN/control
echo "Version:$2" >> /tmp/$1-$2/DEBIAN/control
echo "Section:base" >> /tmp/$1-$2/DEBIAN/control
echo "Priority:optional" >> /tmp/$1-$2/DEBIAN/control
echo "Architecture:amd64" >> /tmp/$1-$2/DEBIAN/control
echo "Depends:"$3 >> /tmp/$1-$2/DEBIAN/control
echo "Maintainer:vejmarie@ruggedpod.qyshare.com" >> /tmp/$1-$2/DEBIAN/control
echo "Homepage:http://ruggedpod.qyshare.com" >> /tmp/$1-$2/DEBIAN/control
echo "Description:TEST PACKAGE" >> /tmp/$1-$2/DEBIAN/control
file_list=`ls -ltd $(find /opt/local/FreeCAD-0.19) | awk '{ print $9}'`
for file in $file_list
do
  is_done=`cat /tmp/stage | grep $file`
if [ "$is_done" == "" ]
then
# The file must be integrated into the new deb
    cp --parents $file /tmp/$1-$2
    echo $file >> /tmp/stage
fi    
done
current_dir=`pwd`
cd /tmp
dpkg-deb --build $1-$2
is_ubuntu=`lsb_release -a | grep -i Distributor | awk '{ print $3}'`
if [ "$is_ubuntu" == "Ubuntu" ]
then
	ubuntuVersion=`lsb_release -a | grep -i codename | awk '{ print $2 }'`
else
	ubuntuVersion="trusty"
fi

mv $1-$2.deb $1-$2_$ubuntuVersion-amd64.deb
cd $current_dir
}

# If we are running ubuntu we must know if we are under Xenial (latest release) or Trusty which
# doesn't support the same package name
if [ "$INIT_DISTRO" == "1" ]
then
	is_ubuntu=`lsb_release -a | grep -i Distributor | awk '{ print $3}'`
	if [ "$is_ubuntu" == "Ubuntu" ]
	then
		ubuntu_version=`lsb_release -a | grep -i codename | awk '{ print $2 }'`
	        if [ "$ubuntu_version" == "xenial" ]
	        then
		add-apt-repository ppa:thopiekar/pyside-git
		apt-get update
		package_list="		doxygen                          \
                              		libboost1.58-dev                 \
                               		libboost-filesystem1.58-dev      \
                               		libboost-program-options1.58-dev \
                              		libboost-python1.58-dev          \
                               		libboost-regex1.58-dev           \
                               		libboost-signals1.58-dev         \
                               		libboost-system1.58-dev          \
                               		libboost-thread1.58-dev          \
                               		libcoin80v5                      \
                               		libcoin80-dev                    \
                               		libeigen3-dev                    \
                               		libpyside-dev                    \
 	                                libqtcore4                       \
                               		libqtcore4                       \
                               		libshiboken-dev                  \
                               		libxerces-c-dev                  \
                               		libxmu-dev                       \
                               		libxmu-headers                   \
                               		libxmu6                          \
                               		libxmuu-dev                      \
                               		libxmuu1                         \
                               		pyside-tools                     \
                               		python-dev                       \
					python3-dev			 \
					python3-numpy			 \
                               		python-pyside                    \
					libpyside2-dev			 \
					python3-pyside2			 \
					pyside2-tools			 \
					libshiboken2-dev		 \
                               		python-matplotlib                \
                               		qt4-dev-tools                    \
                               		qt4-qmake                        \
					libqt5webkit5-dev		\
					libqt5svg5-dev			\
					libqt5xmlpatterns5-dev		\
			       		libqtwebkit-dev			\
                               		shiboken                         \
			       		libcurl4-openssl-dev		\
			       		libssl-dev			\
                               		calculix-ccx                     \
					qttools5-dev			\
                               		swig"
		fi
	fi
	sudo apt-get update
	sudo apt-get install -y dictionaries-common
	sudo apt-get install -y $package_list
	sudo apt-get install -y python-pivy
	sudo apt-get install -y git
	sudo apt-get install -y cmake
	sudo apt-get install -y g++
	sudo apt-get install -y libfreetype6-dev
	sudo apt-get install -y tcl8.5-dev tk8.5-dev
	sudo apt-get install -y libtogl-dev
	sudo apt-get install -y libhdf5-dev
	sudo apt-get install -y xfce4 xfce4-goodies
	sudo apt-get install -y xubuntu-default-settings
	sudo apt-get install -y lightdm
	sudo apt-get install -y automake
	sudo apt-get install -y libcanberra-gtk-module
	sudo apt-get install -y libcanberra-gtk3-module overlay-scrollbar-gtk2 unity-gtk-module-common libatk-bridge2.0-0 unity-gtk2-module libatk-adaptor
	myversion=`lsb_release -a | grep -i Distributor | awk '{ print $3}'`
	if [ "$myversion" != "Ubuntu" ]
	then
	sudo cat /etc/lightdm/lightdm.conf | sed 's/\#autologin-user=/autologin-user=vagrant/' > /tmp/lightdm.conf
	sudo cp /tmp/lightdm.conf /etc/lightdm/lightdm.conf
	sudo /etc/init.d/lightdm start
	sudo apt-get install libmed
	sudo apt-get install libmedc
	sudo apt-get install -y libmed-dev
	sudo apt-get install -y libmedc-dev
	else
cat <<EOF >& /tmp/lightdm.conf
[SeatDefaults]
user-session=xfce
autologin-session=xfce
autologin-user=ubuntu
autologin-user-timeout=0
greeter-session=xfce
pam-service=lightdm-autologin
EOF
	sudo cp /tmp/lightdm.conf /etc/lightdm/
	sudo rm /etc/lightdm/lightdm-gtk-greeter.conf
	sudo /etc/init.d/lightdm start &
	fi 
fi

# Switch boost to the relevant python version
rm /usr/lib/x86_64-linux-gnu/libboost_python.so
ln -s /usr/lib/x86_64-linux-gnu/libboost_python-py35.so /usr/lib/x86_64-linux-gnu/libboost_python.so
# This is to workaround a bug into the Path module
cp /usr/lib/x86_64-linux-gnu/libboost_python-py35.so /usr/lib/x86_64-linux-gnu/libboost_python-py27.so


# Building MED

exist=`ls /tmp/deb/MED*deb`
if [ "$exist" == "" ]
then
	git clone https://github.com/vejmarie/libMED.git
	cd libMED
	mkdir build
	cd build
	cmake -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 ..
	make -j $CPU
	sudo make install
	cd ../..
# I must create the package

	create_deb MED 3.10 ""
	rm -rf libMED
	rm -rf /tmp/MED-3.10

fi

# Building VTK
exist=`ls /tmp/deb/VTK*deb`
if [ "$exist" == "" ]
then
	wget http://www.vtk.org/files/release/7.0/VTK-7.0.0.tar.gz
	gunzip VTK-7.0.0.tar.gz
	tar xf VTK-7.0.0.tar
	rm VTK-7.0.0.tar
	cd VTK-7.0.0
	mkdir build
	cd build
# cmake .. -DVTK_RENDERING_BACKEND=OpenGL
	cmake .. -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 -DVTK_Group_Rendering:BOOL=OFF -DVTK_Group_StandAlone:BOOL=ON -DVTK_RENDERING_BACKEND=None
	make -j $CPU
	sudo make install

	create_deb VTK 7.0 ""
	cd ../..
	rm -rf VTK-7.0.0
fi

# Building OCCT
exist=`ls /tmp/deb/OCC*deb`
if [ "$exist" == "" ]
then
	wget "http://git.dev.opencascade.org/gitweb/?p=occt.git;a=snapshot;h=fd47711d682be943f0e0a13d1fb54911b0499c31;sf=tgz"
	mv "index.html?p=occt.git;a=snapshot;h=fd47711d682be943f0e0a13d1fb54911b0499c31;sf=tgz" occt.tgz
	gunzip occt.tgz
	tar xf occt.tar
	rm occt.tar
	cd occt-fd47711
	grep -v vtkRenderingFreeTypeOpenGL src/TKIVtk/EXTERNLIB >& /tmp/EXTERNLIB
	\cp /tmp/EXTERNLIB src/TKIVtk/EXTERNLIB
	grep -v vtkRenderingFreeTypeOpenGL src/TKIVtkDraw/EXTERNLIB >& /tmp/EXTERNLIB
	\cp /tmp/EXTERNLIB src/TKIVtkDraw/EXTERNLIB
	mkdir build
	cd build
# cmake .. -DUSE_VTK:BOOL=ON
	cmake .. -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 -DUSE_VTK:BOOL=OFF
	sudo make -j $CPU
	sudo make install

	create_deb OCCT 7.4 "vtk (>= 7.0), med (>= 3.10)"
	cd ../..
	rm -rf occt-fd47711
	rm -rf /tmp/OCCT-7.4/
fi

# Building Netgen


exist=`ls /tmp/deb/Net*deb`
if [ "$exist" == "" ]
then
	git clone https://github.com/vejmarie/Netgen
	cd Netgen/netgen-5.3.1
	./configure --prefix=/opt/local/FreeCAD-0.19 --with-tcl=/usr/lib/tcl8.5 --with-tk=/usr/lib/tk8.5  --enable-occ --with-occ=/opt/local/FreeCAD-0.19 --enable-shared --enable-nglib CXXFLAGS="-DNGLIB_EXPORTS -std=gnu++11"
	make -j $CPU
	sudo make install
	cd ../..
	sudo cp -rf Netgen/netgen-5.3.1 /usr/share/netgen
	create_deb Netgen 5.3.1 "occt (>= 7.0)"
	rm -rf Netgen
	rm -rf /tmp/Netgen-5.3.1
fi
# We have to build OpenSSL
#building FreeCAD

wget https://github.com/coin3d/pivy/archive/0.6.5.tar.gz
gunzip 0.6.5.tar.gz
tar xf 0.6.5.tar
cd pivy-0.6.5/
python3 setup_old.py build
sudo python3 setup_old.py install
cd ..
rm -rf pivy-0.6.5 0.6.5.tar


echo "CURRENT DIRECTORY"
pwd
git clone $FREECAD_GIT
cd FreeCAD
git checkout -b $FREECAD_BRANCH origin/$FREECAD_BRANCH
#cat cMake/FindOpenCasCade.cmake | sed 's/\/usr\/local\/share\/cmake\//\/opt\/local\/FreeCAD-0.19\/lib\/cmake/' > /tmp/FindOpenCasCade.cmake
#cp /tmp/FindOpenCasCade.cmake cMake/FindOpenCasCade.cmake
#cp cMake/FindOpenCasCade.cmake cMake/FindOPENCASCADE.cmake
cd ..
mkdir build
cd build
# cmake ../FreeCAD -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 -DBUILD_CLOUD=1 -DALLOW_SELF_SIGNED_CERTIFICATE=1 -DBUILD_FEM=1 -DBUILD_FEM_VTK=1 -DBUILD_FEM_NETGEN=1 -DCMAKE_CXX_FLAGS="-DNETGEN_V5"
cmake ../FreeCAD -DBOOST_PYTHON_SUFFIX=35 -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 -DBUILD_CLOUD=1 -DALLOW_SELF_SIGNED_CERTIFICATE=1 -DBUILD_FEM=1 -DPYTHON_EXECUTABLE=/usr/bin/python3 -DBUILD_QT5=ON -DFREECAD_USE_QWEBKIT:BOOL=ON
make -j $CPU
make -j 4 install
create_deb FreeCAD 0.19 "netgen (>= 5.3.1), occt (>= 7.0), med (>= 3.10)"
cd ..
\rm -rf build
\rm -rf FreeCAD
\rm -rf parts
\rm -rf /tmp/FreeCAD-0.19
source_dir=`pwd`
cd /tmp
mkdir deb
mv *.deb deb
cd deb
dpkg-scanpackages . /dev/null | gzip -9c > Packages.gz
cd ..

pwd
mkdir $source_dir/Results
mv /tmp/deb/*.deb $source_dir/Results
current_dir=`pwd`
cd $source_dir/Results
dpkg-scanpackages . /dev/null > Release
dpkg-scanpackages . /dev/null | gzip -9c > Packages.gz
cd $current_dir
echo "deb [trusted=yes] file://$source_dir/Results /" >> /etc/apt/sources.list
apt-get update

if [ "$ubuntu_version" == "xenial" ]
then
#Let's build the snap
	export LC_ALL="en_US.UTF-8"
	export LANG="en_US.UTF-8"
	export LANGUAGE="en_US.UTF-8"
	mkdir snap
	cd snap
	cp -Rf /vagrant/* .
        apt-get install -y snapcraft
#	./generate_yaml.sh
	snapcraft
	mv freecad_0.19_amd64.snap /tmp
fi
if [ -f "/tmp/freecad_0.19_amd64.snap" ]
then
	mv /tmp/freecad_0.19_amd64.snap $source_dir/Results
fi

# cmake ../FreeCAD -DCMAKE_INSTALL_PREFIX:PATH=/opt/local/FreeCAD-0.19 -DBUILD_CLOUD=1 -DALLOW_SELF_SIGNED_CERTIFICATE=1 -DBUILD_FEM=1 -DPYTHON_EXECUTABLE=/usr/bin/python3 -DBUILD_QT5=ON -DFREECAD_USE_QWEBKIT:BOOL=ON
