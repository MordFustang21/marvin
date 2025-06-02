#import <Cocoa/Cocoa.h>
#include <stdlib.h>

void FreeIconData(void *data) {
    free(data);
}

// Returns PNG image data for the icon of a file path
// The caller must free the returned buffer using FreeIconData
unsigned char *GetAppIconForPath(const char *path, int *length) {
    @autoreleasepool {
        NSString *nsPath = [NSString stringWithUTF8String:path];
        NSImage *icon = [[NSWorkspace sharedWorkspace] iconForFile:nsPath];
        if (!icon) {
            *length = 0;
            return NULL;
        }

        // Convert to PNG representation
        NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithData:[icon TIFFRepresentation]];
        NSData *pngData = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];

        if (!pngData) {
            *length = 0;
            return NULL;
        }

        *length = (int)[pngData length];
        unsigned char *buffer = (unsigned char *)malloc(*length);
        memcpy(buffer, [pngData bytes], *length);
        return buffer;
    }
}
