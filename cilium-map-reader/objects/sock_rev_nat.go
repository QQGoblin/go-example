package objects

import (
	"fmt"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/byteorder"
	"github.com/cilium/cilium/pkg/types"
	"unsafe"
)

type SockRevNat4Key struct {
	Cookie  uint64     `align:"cookie"`
	Address types.IPv4 `align:"address"`
	Port    uint16     `align:"port"`
	Pad     uint16     `align:"pad"`
}

type SockRevNat4Value struct {
	Address     types.IPv4 `align:"address"`
	Port        uint16     `align:"port"`
	RevNatIndex uint16     `align:"rev_nat_index"`
}

func (in *SockRevNat4Key) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(in) }

func (in *SockRevNat4Key) String() string {
	h := *in
	h.Port = byteorder.NetworkToHost16(h.Port)
	return fmt.Sprintf("[%s]:%d, %d", in.Address.String(), h.Port, in.Cookie)
}
func (in SockRevNat4Key) NewValue() bpf.MapValue { return &SockRevNat4Value{} }

func (in *SockRevNat4Value) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(in) }

func (in *SockRevNat4Value) String() string {
	h := *in
	h.Port = byteorder.NetworkToHost16(h.Port)
	return fmt.Sprintf("[%s]:%d, %d", in.Address.String(), h.Port, in.RevNatIndex)
}
