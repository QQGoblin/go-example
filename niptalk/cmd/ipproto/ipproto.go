package ipproto

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"syscall"
)

// MarshalIPPkt 通过 github.com/google/gopacket 提供的工具，可以快速生成带 IP Header 的数据包
func MarshalIPPkt(srcIP, dstIP string, protocal layers.IPProtocol, payload []byte) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	var layersToSerialize []gopacket.SerializableLayer

	ipLayer := &layers.IPv4{
		Version: 4,
		TTL:     64,
		SrcIP:   net.ParseIP(srcIP),
		// 当我们使用系统调用 Sendto 发送数据包时 SockaddrInet4 中填写的目的地址会自动填充到 IP 头
		DstIP:    net.ParseIP(dstIP),
		Protocol: protocal,
	}

	layersToSerialize = append(layersToSerialize, ipLayer)
	layersToSerialize = append(layersToSerialize, gopacket.Payload(payload))

	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}, layersToSerialize...); err != nil {
		return nil, errors.Wrapf(err, "Error serializing packet")
	}

	return buf.Bytes(), nil
}

func SendIPPkt(srcIP, dstIP string, protocol int, payload []byte) error {

	buf, err := MarshalIPPkt(srcIP, dstIP, layers.IPProtocol(protocol), payload)
	if err != nil {
		return errors.Wrapf(err, "Failed to Marshal IPv4 Packet")
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		return errors.Wrapf(err, "Failed to create raw socket")
	}
	defer syscall.Close(fd)

	// 可以通过设置`IP_HDRINCL`参数自定义网络包的`IP Header`信息（PS：未指定该参数时，系统会自动填充`IP`头）
	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		return errors.Wrapf(err, "Failed to set IP_HDRINCL")
	}

	dstAddr := net.ParseIP(dstIP).To4()

	addr := &syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{dstAddr[0], dstAddr[1], dstAddr[2], dstAddr[3]},
	}

	// 参考 https://man7.org/linux/man-pages/man3/sendto.3p.html
	return syscall.Sendto(fd, buf, 0, addr)
}

func ReceiveIPPkt(protocol int) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, protocol)
	if err != nil {
		return errors.Wrapf(err, "syscall socket failed")
	}
	fmt.Printf("Obtained fd %d", fd)
	defer syscall.Close(fd)

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))
	defer f.Close()

	for {
		buf := make([]byte, 1500) // Base on mtu
		numRead, err := f.Read(buf)
		if err != nil {
			return errors.Wrapf(err, "problems @ location 2")
		}

		// 解析从原始报文中解析 IP 报文头
		ipHeader, err := ipv4.ParseHeader(buf[:numRead])
		if err != nil {
			return errors.Wrapf(err, "problems parase ipv4 header")
		}

		fmt.Printf("Receive %s -> %s: %s\n", ipHeader.Src.String(), ipHeader.Dst.String(), buf[ipv4.HeaderLen:])
		// 解析 TCP 报文头，"gvisor.dev/gvisor/pkg/tcpip/header"
		// tcpPayload := header.TCP(buf[ipv4.HeaderLen:])
		// fmt.Printf("Receive %s:%d -> %s:%d\n", ipHeader.Src.String(), tcpPayload.SourcePort(), ipHeader.Dst.String(), tcpPayload.DestinationPort())
	}
}
