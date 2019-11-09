# SAPmon
SAPmon reads in output of CCMS function calls on a SAP system and converts it to metrics consumable by telegrafs exec input plugin.

## Get it

Use the makefile from the repository:

```
make all
```

You could also build only for specific operating system:

```
make build_windows
make build_linux
```

## Usage

Currently there are the following function call outputs supported:
* GetAlertTree
* GetProcessList

Issue the following command as *SIDadm* on the SAP system:

```
sapcontrol -nr 00 -function GetAlertTree | ./sapmon
```

It prints out all metrics on STDOUT. This could be used in telegrafs exec input plugin:

```
[[inputs.exec]]
  commands = ["sudo su - SIDadm -c 'sapcontrol -nr 00 -function GetAlertTree | /usr/local/bin/sapmon'"]
  timeout = "30s"
  data_format = "influx"
```

*Adjust sudoers if necessary!*

## Command Line Flags

**general flags:**
* **-m**: Metric measurement name. Default: sapmon
* **-r**: Replace spaces in tag keys/values and field keys with delimiter, see also *parameter d*. Default: false
* **-d**: Delimiter to separate node names and replace spaces in tag keys/values and field keys, see also *parameter r*. Default: space
* **-p**: Name of tag where *node path (GetAlertTree)* **or** *process list (GetProcessList* is stored. If set to ROOT the name of the root node will be used. Default: ccms
* **-v**: Verbose output. Default: false

**GetAlertTree function only:**
* **-n**: Number of node which is our root to start from. Default: 0
* **-t**: Use timestamp from SAP AlertNode for metric. Default: false
* **-u**: Field keys for which the SAP AlertNode timestamp should be set. Options: string, float, all = other input. Default: all

See **./sapmon -h** for more details.

## License
[Apache License 2.0](https://github.com/phlipse/sapmon/blob/master/LICENSE)
