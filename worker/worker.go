package worker

import (
    "fmt"
    "bytes"
    "encoding/json"
    "golang.org/x/crypto/ssh"
    libvirt "github.com/libvirt/libvirt-go"
)

type Domain struct {
    Name string `xml:"name"`
    Memory int `xml:"memory"`
    Vcpu int `xml:"vcpu"`
    Devices struct {
        Channel struct {
            Target struct {
                Type string `xml:"type"`
                Name string `xml:"type"`
            } `xml:"target"`
        } `xml:"channel"`
        Interface struct {
            Source struct {
                Bridge string `xml:"bridge"`
            } `xml:"source"`
            VirtualPort struct {
                Type string `xml:"type"`
            } `xml:"virtualport"`
        } `xml:"interface"`
        Disk struct {
            Driver struct {
                Type string `xml:"type,attr"`
                Cache string `xml:"cache,attr"`
            } `xml:"driver"`
            Source struct {
                File string `xml:"file,attr"`
            } `xml:"source"`
            Target struct {
                Dev string `xml:"dev,attr"`
                Bus string `xml:"bus,attr"`
            } `xml:"target"`
        } `xml:"disk"`
    } `xml:"devices"`
    Os struct {
        Type string `xml:"type"`
        Boot string `xml:"boot"`
    } `xml:"os"`
    Features struct {
        Acpi string `xml:"acpi"`
    } `xml:"features"`
}

type QemuAgentCommandRequest struct {
    Execute string `json:"execute"`
}

type NetworkInterface struct {
    Prefix uint32 `json:"prefix"`
    IpAddress string `json:"ip-address"`
    IpAddressType string `json:"ip-address-type"`
}

type NetworkInterfaces struct {
    Name string `json:"name"`
    HardwareAddress string `json:"hardware-address"`
    IpAddresses []NetworkInterface `json:"ip-addresses"`
}

type QemuAgentCommandResponse struct {
    Return []NetworkInterfaces `json:"return"`
}

func Work() {
    fmt.Println("fuck")
    conn, err := libvirt.NewConnect("qemu:///system")
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    command := &QemuAgentCommandRequest{
        Execute: "guest-network-get-interfaces",
    }
    jsonCommand, _ := json.Marshal(command)
    doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_SHUTOFF)
    if err != nil {
        panic(err)
    }
    sshConfig := &ssh.ClientConfig{
        User: "ubuntu",
        Auth: []ssh.AuthMethod{
            ssh.Password("passw0rd"),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    for _, dom := range doms {
        dm, err := dom.GetName()
        fmt.Println(dm)
        dom.Create()

        name, err := dom.QemuAgentCommand(string(jsonCommand), libvirt.DOMAIN_QEMU_AGENT_COMMAND_MIN, 0)
        for err != nil {
            name, err = dom.QemuAgentCommand(string(jsonCommand), libvirt.DOMAIN_QEMU_AGENT_COMMAND_MIN, 0)
        }
        var keys QemuAgentCommandResponse
        json.Unmarshal([]byte(name), &keys)
        if err == nil {
            fmt.Printf("%s\n", name)
        } else {
            panic(err)
        }
        var buffer bytes.Buffer
        buffer.WriteString(keys.Return[1].IpAddresses[0].IpAddress)
        buffer.WriteString(":22")
        connection, err := ssh.Dial("tcp", buffer.String(), sshConfig)
        for err != nil {
            connection, err = ssh.Dial("tcp", buffer.String(), sshConfig)
        }
        session, err := connection.NewSession()
        err = session.Run("ls -la > fuckme.txt")
        fmt.Println("fuck")
        dom.Free()
    }
}