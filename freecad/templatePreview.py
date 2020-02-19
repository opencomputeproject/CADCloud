import FreeCAD as App
import FreeCADGui as Gui
import Part, PartGui
import Cloud
import time

doc=FreeCAD.newDocument("NAME")
Gui.SendMsgToActiveView("ViewFit")
myView=Gui.activeDocument().activeView()

Gui.runCommand('Std_PerspectiveCamera',1)
Gui.activeDocument().activeView().setCameraType("Perspective")
Cloud.cloudurl(u"URI")
Cloud.cloudtcpport(u"PORT")
Cloud.cloudaccesskey(u"KEY")
Cloud.cloudsecretkey(u"SECRETKEY")
Cloud.cloudrestore(u"BUCKET")
from pivy import coin
pos3=FreeCAD.Vector(0,0,0)
camera = FreeCADGui.ActiveDocument.ActiveView.getCameraNode()
campos=FreeCAD.Vector(1,1,1)
camera.position.setValue( campos)
camera.pointAt(coin.SbVec3f(pos3),coin.SbVec3f(0,0,1))
myView.fitAll()


Gui.ActiveDocument.ActiveView.setAnimationEnabled(False)
par=FreeCAD.ParamGet("User parameter:BaseApp")
grp=par.GetGroup("Preferences/View")

#grp.SetBool("DisablePBuffers",False)

grp.SetString("SavePicture","FramebufferObject")
myView.saveImage("FILE_PATH",1920,1080,"NAME")
Gui.runCommand('Std_SelectAll',0)
__objs__=[]
for obj in FreeCAD.ActiveDocument.Objects:
 __objs__.append(FreeCAD.ActiveDocument.getObject(str(obj.Name)))

import importOBJ
importOBJ.export(__objs__,u"OBJ_PATH")
del __objs__

Gui.doCommand('exit()')
