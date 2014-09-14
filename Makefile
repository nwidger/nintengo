all:
	cp nintengo Nintengo.app/Contents/MacOS && install_name_tool \
	-change /usr/local/lib/libSDL_image-1.2.0.dylib @executable_path/libSDL_image-1.2.0.dylib \
	-change /usr/local/opt/sdl/lib/libSDL-1.2.0.dylib @executable_path/libSDL-1.2.0.dylib \
	-change /usr/local/lib/libGLEW.1.10.0.dylib @executable_path/libGLEW.1.10.0.dylib \
	Nintengo.app/Contents/MacOS/nintengo
