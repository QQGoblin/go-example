package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
	"os/exec"
	"syscall"
)

const (
	vxlanDeviceName = "vxlan100"
	vxlanID         = 468
	vxlanDstPort    = 4789
	defaultMTU      = 1450
)

var (
	underlay  string
	virtualIP string
	neighIPs  []string
	routeIPs  []string
)

func ensureLink(link *netlink.Vxlan) (netlink.Link, error) {

	err := netlink.LinkAdd(link)
	if err == syscall.EEXIST {
		klog.V(1).Infof("VXLAN device already exists")
		existing, err2 := netlink.LinkByName(link.Name)
		if err2 != nil {
			return nil, err2
		}

		if err = netlink.LinkDel(existing); err != nil {
			return nil, fmt.Errorf("failed to delete interface: %v", err)
		}

		if err = netlink.LinkAdd(link); err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	return netlink.LinkByName(link.Name)

}

func localAddr(link netlink.Link) (*netlink.Addr, error) {
	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	var local *netlink.Addr
	for _, addr := range addrs {
		if (addr.Flags & (unix.IFA_F_SECONDARY | unix.IFA_F_DEPRECATED)) != 0 {
			continue
		}
		local = &addr
	}
	if local == nil {
		return nil, fmt.Errorf("not ip address found")
	}

	return local, nil
}

func configure(link netlink.Link, addr *netlink.Addr, neighIPs []string, staticRouteIPs []string) error {

	// 配置虚拟网卡的 IP 地址
	if err := netlink.AddrAdd(link, addr); err != nil {
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set interface %s to UP state: %s", link.Attrs().Name, err)
	}

	// 创建 fdb 记录
	for _, neighIP := range neighIPs {

		cmd := exec.Command(
			"bash", "-c",
			fmt.Sprintf("bridge fdb append to 00:00:00:00:00:00 dst %s dev %s", neighIP, link.Attrs().Name),
		)
		if _, err := cmd.Output(); err != nil {
			return err
		}
	}

	// 创建静态路由
	for _, routeIP := range staticRouteIPs {

		dst, err := netlink.ParseIPNet(fmt.Sprintf("%s/32", routeIP))
		if err != nil {
			return err
		}
		route := netlink.Route{Dst: dst, LinkIndex: link.Attrs().Index}
		if err = netlink.RouteAdd(&route); err != nil {
			return err
		}
	}

	return nil
}

func generateSubnet(iface, virtualIP string, neighIPs, routeIPs []string) error {

	// 获取 link
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	// 获取本地 IP 地址
	local, err := localAddr(link)
	if err != nil {
		return err
	}

	// 创建 vxlan 设备
	vxlan := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name: vxlanDeviceName,
			MTU:  defaultMTU,
		},
		VxlanId:      vxlanID,
		VtepDevIndex: link.Attrs().Index,
		SrcAddr:      local.IP,
		Port:         vxlanDstPort,
		Learning:     true,
	}

	vxlanLink, err := ensureLink(vxlan)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/32", virtualIP))
	if err = configure(vxlanLink, addr, neighIPs, routeIPs); err != nil {
		return err
	}

	return nil
}
