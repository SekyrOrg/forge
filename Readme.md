# BeaconForge

BeaconForge is a powerful and versatile tool for generating beacons from file paths.
It is designed to simplify the process of creating and deploying beacons,
allowing you to easily integrate them into your system. With BeaconForge,
you can fine-tune the behavior of your beacon to suit your specific needs,
giving you complete control over the functionality and performance of your beacon.

### Usage
```
BeaconServer is a server that listens for beacon events and sends them to the backend server.

Usage:
  ./main [flags]

Flags:
OPTIONS FOR THE BEACONS:
   -c, -connectionString string  Connection string for the beacon server (default "127.0.0.1:5353")
   -t, -transport string         Transport to use for the beacon server (http, dns, tcp) (default "http")
   -s, -static                   Build a static binary
   -u, -upx                      Compress the binary with upx
   -ul, -upxLevel int            Compression level for upx (default 1)
   -d, -debug                    Enable beacon debug output

OTHER OPTIONS:
   -a, -addr string     Address of the beacon creator server (default "http://127.0.0.1:8080")
   -f, -files string[]  FilePaths to be used as beacon
   -v, -verbose         Enable verbose output
   -C, -config string   Path to the configuration file

```

One of the key features of BeaconForge is its wide range of options for customizing
the behavior of your beacon.These options include the ability to specify a different
connection string,transport protocol, and compression level, giving you full control
over how your beacon communicates with the backend. Additionally, BeaconForge allows
you to enable verbose output,providing detailed information about the operation of
your beacon and helping you to diagnose any issues that may arise.

Another important feature of BeaconForge is its support for configuration files. 
By specifying a configuration file, you can streamline your workflow and
simplify the process of creating and deploying your beacon.
This can be especially useful if you are working with large numbers of files
or need to repeat the process of creating a beacon multiple times.

In addition to its powerful features, BeaconForge is also designed to be easy to use.
With its straightforward command-line interface, you can quickly and easily generate
beacons from file paths, and customize their behavior to suit your specific needs.
Whether you're an experienced developer or just getting started with beacon technology,
BeaconForge is the ideal tool for creating and deploying high-performance, reliable beacons.

In conclusion, BeaconForge is a powerful and versatile tool that makes it easy to create
and deploy beacons. With its wide range of options for customizing the behavior of your
beacon and its support for configuration files, it gives you complete control over the
functionality and performance of your beacon. Additionally, its easy-to-use command-line
interface makes it accessible to developers of all skill levels. If you're looking for
a reliable, easy-to-use solution for creating beacons, BeaconForge is the perfect choice.
