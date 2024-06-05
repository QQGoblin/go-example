package objects

import (
	"fmt"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/byteorder"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"strings"
)

func flagsString(c *ctmap.CtEntry) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Flags=%#04x [ ", c.Flags))
	if (c.Flags & ctmap.RxClosing) != 0 {
		sb.WriteString("RxClosing ")
	}
	if (c.Flags & ctmap.TxClosing) != 0 {
		sb.WriteString("TxClosing ")
	}
	if (c.Flags & ctmap.Nat64) != 0 {
		sb.WriteString("Nat64 ")
	}
	if (c.Flags & ctmap.LBLoopback) != 0 {
		sb.WriteString("LBLoopback ")
	}
	if (c.Flags & ctmap.SeenNonSyn) != 0 {
		sb.WriteString("SeenNonSyn ")
	}
	if (c.Flags & ctmap.NodePort) != 0 {
		sb.WriteString("NodePort ")
	}
	if (c.Flags & ctmap.ProxyRedirect) != 0 {
		sb.WriteString("ProxyRedirect ")
	}
	if (c.Flags & ctmap.DSR) != 0 {
		sb.WriteString("DSR ")
	}

	unknownFlags := c.Flags
	unknownFlags &^= ctmap.MaxFlags - 1
	if unknownFlags != 0 {
		sb.WriteString(fmt.Sprintf("Unknown=%#04x ", unknownFlags))
	}
	sb.WriteString("]")
	return sb.String()
}

func CT4DumpCallback(key bpf.MapKey, value bpf.MapValue) {

	ctKey4 := key.(*ctmap.CtKey4Global)
	ctEntry := value.(*ctmap.CtEntry)

	var sb strings.Builder
	if !ctKey4.ToHost().Dump(&sb, true) {
		return
	}

	backendID := -1
	if ctKey4.Flags&ctmap.TUPLE_F_SERVICE != 0 {
		backendID = int(ctEntry.RxPackets)
	}

	sb.WriteString(
		fmt.Sprintf("BackendID=%d RxFlagsSeen=%#02x LastRxReport=%d TxFlagsSeen=%#02x LastTxReport=%d %s RevNAT=%d SourceSecurityID=%d IfIndex=%d",
			backendID,
			ctEntry.RxFlagsSeen,
			ctEntry.LastRxReport,
			ctEntry.TxFlagsSeen,
			ctEntry.LastTxReport,
			flagsString(ctEntry),
			byteorder.NetworkToHost16(ctEntry.RevNAT),
			ctEntry.SourceSecurityID,
			ctEntry.IfIndex))
	fmt.Println(sb.String())

}
