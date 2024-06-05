package objects

import (
	"fmt"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/loadbalancer"
	"github.com/cilium/cilium/pkg/maps/lbmap"
	"net"
)

func LBServiceDumpCallback(key bpf.MapKey, value bpf.MapValue) {

	k := key.(*lbmap.Service4Key)
	kHost := k.ToHost().(*lbmap.Service4Key)
	s := value.(*lbmap.Service4Value)
	addr := net.JoinHostPort(k.Address.String(), fmt.Sprintf("%d", kHost.Port))
	if kHost.Scope == loadbalancer.ScopeInternal {
		addr += "/i"
	}

	addr += fmt.Sprintf(" [%d]", k.BackendSlot)

	fmt.Printf("%-20s %d %d (%d) [0x%x 0x%x]\n", addr, s.BackendID, s.Count, s.RevNat, s.Flags, s.Flags2)

}
