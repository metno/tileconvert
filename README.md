# Tileconvert - sample code for modifying tileserver images server-side

This contains sample code for a web server, acting as a proxy for a tileserver.
The proxy modifies the sent image tiles before passing them on, for example by changing the colors in the image.

**⚠ Note:** This code is not meant for production use! Rather, it is an example to get you started when creating your own tile converter.

The most relevant code to read is found [here](cloudconverter.go#L27).
