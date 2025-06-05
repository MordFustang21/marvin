package macos

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

void setDockIconVisibility(bool visible) {
    if (visible) {
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
    } else {
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    }
}

void* getObjcClass(const char* className) {
    return (void*)objc_getClass(className);
}

void* registerSelector(const char* selectorName) {
    return (void*)sel_registerName(selectorName);
}

void* sendMessage(void* target, void* selector, int arg) {
    return (void*)[(id)target performSelector:(SEL)selector withObject:(id)(NSInteger)arg];
}

void* sendSimpleMessage(void* target, void* selector) {
    return (void*)[(id)target performSelector:(SEL)selector];
}
*/
import "C"
import (
	"unsafe"
)

// HideDockIcon hides the application from the dock
func HideDockIcon() {
	C.setDockIconVisibility(false)
}

// ShowDockIcon shows the application in the dock
func ShowDockIcon() {
	C.setDockIconVisibility(true)
}

// ObjectiveC runtime wrappers for use in other packages

// ObjcGetClass gets an Objective-C class by name
func ObjcGetClass(className string) unsafe.Pointer {
	cstr := C.CString(className)
	defer C.free(unsafe.Pointer(cstr))
	return C.getObjcClass(cstr)
}

// SelRegisterName registers a selector with the Objective-C runtime
func SelRegisterName(selectorName string) unsafe.Pointer {
	cstr := C.CString(selectorName)
	defer C.free(unsafe.Pointer(cstr))
	return C.registerSelector(cstr)
}

// ObjcMsgSend sends a message to an Objective-C object
func ObjcMsgSend(target unsafe.Pointer, selector unsafe.Pointer, arg int) unsafe.Pointer {
	return C.sendMessage(target, selector, C.int(arg))
}

// ObjcMsgSendSimple sends a message to an Objective-C object with no arguments
func ObjcMsgSendSimple(target unsafe.Pointer, selector unsafe.Pointer) unsafe.Pointer {
	return C.sendSimpleMessage(target, selector)
}
