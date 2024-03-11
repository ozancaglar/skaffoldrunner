Run skaffold configurations.

To install, ensure you have go installed and...

`go install github.com/ozancaglar/skaffoldrunner`

Flags

--file -f Path to `skaffold.yaml`, if no file is specified, then it assumes there is one in `pwd` and that you want to run all of the modules.

--workdir -w Path to where you want to run `skaffold`, defaults to `pwd`

If your Go is in your PATH then you should be able to...

`skaffoldrunner -f skaffold.yaml`
