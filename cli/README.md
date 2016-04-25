kode
=========

Kode is koderunr's command line client, which mean's you can run the code locally even you don't have the code installed locally. Currently heavily under development.

Developing kode
--------------------

Make sure that you have go and ruby installed. Install koderunr into your `GOPATH` by `go get github.com/jaxi/koderunr`, then change your directory to `$GOPATH/src/github.com/jaxi/koderunr/client`

Kode command line tool uses `rake` to cross compile the source code (so far didn't find a better way doing it). To do it, run

```
rake cross_compile
```

See `rake -T` For more flexibility.

Then you can get the `kode`'s binary in build folder according to the OS and architecture of your computer.

Distribution
-------

TODO
