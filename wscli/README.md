wscli is a simple WSMAN command line tool.  Right now, it has support for
the following features:

* WSMAN Get, Put, Create, Delete, Invoke, Enumerate, and EnumerateEPR.
* HTTP and HTTPS transports, using Basic auth.
* Enumerate always optimizes and pulls the complete result set.
* Put and Create accept XML input on stdin.


wscli is just a thin wrapper around github.com/VictorLowther/wsman.  As
that library gains features, so will wscli.

Build instructions:

    go get github.com/VictorLowther/wsman
    cd $GOPATH/src/github.com/VictorLowther/wsman/wscli
    go build

Usage examples:

Identify a WSMAN endpoint:

    wscli -e https://wsman.endpoiint/wsman \
        -u "user" -p 'password'

Turn a Dell Poweredge R720 on:

    wscli -e https://192.168.128.41:443/wsman \
        -u "root" -p 'password' -a Invoke \
        -r http://schemas.dell.com/wbem/wscim/1/cim-schema/2/DCIM_CSPowerManagementService \
        -m RequestPowerStateChange \
        -s "Name: pwrmgtsvc:1, CreationClassName: DCIM_CSPowerManagementService,
            SystemCreationClassName: DCIM_SPComputerSystem, SystemName: systemmc" \
        -x "PowerState: 2"

Exit codes on failure:

1. SOAP Fault message returned
2. Transport error
3. Argument syntax error
