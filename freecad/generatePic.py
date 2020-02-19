import FreeCAD as App
import FreeCADGui as Gui
import Part, PartGui
import Cloud
import time

print "doc\n"
doc=FreeCAD.newDocument("titi")
#console.log(doc)
print "doc"
Gui.SendMsgToActiveView("ViewFit")
myView=Gui.activeDocument().activeView()
Gui.runCommand('Std_PerspectiveCamera',1)
Gui.activeDocument().activeView().setCameraType("Perspective")
box = doc.addObject("Part::Box","myBox")
box.Height=4
box.Width=2
doc.recompute()
#Cloud.cloudurl(u"https://justyour.parts")
#Cloud.cloudtcpport(u"443")
#Cloud.cloudaccesskey(u"")
#Cloud.cloudsecretkey(u"")
#Cloud.cloudrestore(u"test")
#box = doc.addObject("Part::Box","myBox")
#box.Height=4
#box.Width=2
#doc.recompute()

from pivy import coin
pos3=FreeCAD.Vector(0,0,0)
camera = FreeCADGui.ActiveDocument.ActiveView.getCameraNode()
campos=FreeCAD.Vector(1,1,1)
camera.position.setValue( campos)
camera.pointAt(coin.SbVec3f(pos3),coin.SbVec3f(0,0,1))
myView.fitAll()

myView.fitAll()
#Gui.runCommand('Std_PerspectiveCamera',1)
Gui.ActiveDocument.ActiveView.setAnimationEnabled(False)
par=FreeCAD.ParamGet("User parameter:BaseApp")
grp=par.GetGroup("Preferences/View")
#grp.SetBool("DisablePBuffers",False)
grp.SetString("SavePicture","FramebufferObject")
myView.saveImage('/home/ubuntu/test.png',54,28,'titi')
