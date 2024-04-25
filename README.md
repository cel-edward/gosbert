# SBERT Python functionality called from Go

This library is designed to expose simple functionalities from the popular Python library sentence-transformers into Go, given the lack of local implementation.

This is strongly a work-in-progress in very early developmental stages, and is not suitable for external use in its current form.c

# Setup
## cpy3
Must have Python 3.8 and associated headers (python3.8-dev) installed. On Ubuntu:

    sudo add-apt-repository ppa:deadsnakes/ppa
    sudo apt install python3.8
    sudo apt-get install python3.8-dev

Then must set pkg-config. Easiest to copy and store locally. On Ubuntu:
- Get from `/usr/lib/x86_64-linux-gnu/pkgconfig/python-3.8-embed.pc`
- Copy to `pkg-config` and rename as python3.pc
- Run the commands in `set_env.sh`

If done correctly `echo $PKG_CONFIG_PATH` should print the pkg-config directory in the module folder

If something goes wrong during installation / usage you may need to reinstall with clean caches:

    go clean github.com/cel-edward/cpy3
    go clean -cache
    (run set_env.sh again)
    go get github.com/cel-edward/cpy3
    

## Python

If not already installed: `sudo apt install python3.8-distutils`

Then install dependencies (slimmed version of sentence-transformers)
    
    python3.8 -m pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu
    python3.8 -m pip install --no-cache-dir sentence-transformers
    python3.8 -m pip install --no-cache-dir Pillow --upgrade


# Usage

Tests must be compiled and run from the working directory where you would otherwise run your builds.

    go test -c && ./gosbert.test


# References
https://poweruser.blog/embedding-python-in-go-338c0399f3d5

https://github.com/DataDog/go-python3/issues/38

https://stackoverflow.com/questions/77205123/how-do-i-slim-down-sberts-sentencer-transformer-library

