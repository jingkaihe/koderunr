# Koderunr

Koderunr (read code runner) is a container-based service that allows you to run code without programming language installation pain, instead the program will be executed remotely in a container.

![Gif Example](http://g.recordit.co/GY2xkSf16B.gif)

## Install

You can go to the cli directory and run `go build` to get the command line binary.

Or you can cross compile to the binaries that running on platforms including OS X, Linux, FreeBSD and OpenBSD by doing `rake build`, which can allow you to build production ready binaries.

## Web Interface

Now it's live on http://koderunr.tech!

## Examples

Suppose you have a ruby file called `foo.rb`, which has the source

```ruby
print "Enter a number to count down: "
$stdout.flush
num = readline.to_i rescue 0

num.downto(0) do |i|
  puts "...#{i}"
  $stdout.flush
  sleep 1
end

puts "boom!"
```

You can execute the source code by running the command - and the results will be outputted as below with time intervals.

```bash
$ kode run foo.rb
Enter a number to count down: 4
...4
...3
...2
...1
...0
boom!
```

Also you can specified version of the language you want by

```bash
$ kode run foo.rb -version=2.3.0 # Running foo.rb using ruby 2.2.0
```

## TODO

- [x] ~~Support more languages (e.g. C, python, ruby, Erlang), at the moment only Go is supported.~~ Now supporting Go, C, ruby, python.
- [ ] Running the Docker containers in a proper Docker client rather than using system calls, so we can create-attach-run-kill containers automatically. (Currently tidy up the dead containers manually)
- [x] Serious cli support
- [x] cli configuration
- [x] Binaries distribution
- [x] Web interface that sucks less...
- [x] Server and Docker containers hosting
- [x] Support programming language versioning
