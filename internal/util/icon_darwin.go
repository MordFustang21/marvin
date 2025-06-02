package util

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

extern unsigned char *GetAppIconForPath(const char *path, int *length);
extern void FreeIconData(void *data);
*/
import "C"

import (
	"fmt"
	"path/filepath"
	"sync"
	"unsafe"

	"fyne.io/fyne/v2"
)

var (
	iconCache = sync.Map{}
)

func getIconUsingCocoa(path string) (fyne.Resource, error) {
	value, ok := iconCache.Load(path)
	if ok {
		return value.(fyne.Resource), nil
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var length C.int
	data := C.GetAppIconForPath(cPath, &length)
	if data == nil || length == 0 {
		return nil, fmt.Errorf("no icon returned from Cocoa")
	}
	defer C.FreeIconData(unsafe.Pointer(data))

	// Convert to Go byte slice
	bytes := C.GoBytes(unsafe.Pointer(data), length)
	name := filepath.Base(path) + "_icon"

	resource := fyne.NewStaticResource(name, bytes)

	// Store in cache
	iconCache.Store(path, resource)

	return resource, nil
}
