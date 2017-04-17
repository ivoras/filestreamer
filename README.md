# File streamer

This is a simple utility written in Go which monitors a given directory and for each new file found, opens it and streams it to a remote server over TCP. When the EOF is hit, the utility will wait a bit and then read any new data (if it appears in the file) and stream it also.

The utility was written to stream multimedia files as they are being written.

# Connection header

As soon as a connection is established, this utility will send a JSON header (terminated with ``\n``), which is a dictionary containing the following fields:

* ``filename`` is the file name of the file
* ``timestamp`` is the original file timestamp (mtime), when first seen.

After the ``\n`` comes raw file data.