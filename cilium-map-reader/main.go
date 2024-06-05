package main

import (
	"flag"
	"fmt"
	"github.com/QQGoblin/cilium-map-reader/objects"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"github.com/cilium/cilium/pkg/maps/lbmap"
	"github.com/cilium/cilium/pkg/maps/neighborsmap"
	"k8s.io/klog/v2"
	"reflect"
)

type DumpMeta struct {
	MapKey   bpf.MapKey
	MapValue bpf.MapValue
	Parser   bpf.DumpParser
	Callback bpf.DumpCallback
}

const (
	defaultKVFormate = "%-40s %s\n"
)

var (
	bpfMapName = ""

	dumpMeta = map[string]DumpMeta{
		lbmap.Service4MapV2Name:  {MapKey: &lbmap.Service4Key{}, MapValue: &lbmap.Service4Value{}, Callback: objects.LBServiceDumpCallback},
		lbmap.Backend4MapV2Name:  {MapKey: &lbmap.Backend4Key{}, MapValue: &lbmap.Backend4Value{}},
		lbmap.RevNat4MapName:     {MapKey: &lbmap.RevNat4Key{}, MapValue: &lbmap.RevNat4Value{}},
		lbmap.SockRevNat4MapName: {MapKey: &objects.SockRevNat4Key{}, MapValue: &objects.SockRevNat4Value{}},
		ctmap.MapNameTCP4Global:  {MapKey: &ctmap.CtKey4Global{}, MapValue: &ctmap.CtEntry{}, Callback: objects.CT4DumpCallback},
		ctmap.MapNameAny4Global:  {MapKey: &ctmap.CtKey4Global{}, MapValue: &ctmap.CtEntry{}, Callback: objects.CT4DumpCallback},
		neighborsmap.Map4Name:    {MapKey: &neighborsmap.Key4{}, MapValue: &neighborsmap.Value{}},
	}
)

func init() {

	flag.StringVar(&bpfMapName, "name", lbmap.Service4MapV2Name, fmt.Sprintf("bpfmap name %v", reflect.ValueOf(dumpMeta).MapKeys()))

}

func main() {

	flag.Parse()
	dm, isOK := dumpMeta[bpfMapName]
	if !isOK {
		klog.Fatalf("%s dump meta is not registry\n", bpfMapName)
	}

	m, err := bpf.OpenMap(bpfMapName)
	if err != nil {
		klog.Fatalf("open %s failed: %v\n", bpfMapName, err)
	}

	m.MapInfo.MapKey = dm.MapKey
	m.MapInfo.MapValue = dm.MapValue
	m.DumpParser = dm.Parser

	if dm.Parser == nil {
		m.DumpParser = bpf.ConvertKeyValue
	}

	callback := dm.Callback
	if callback == nil {
		callback = func(key bpf.MapKey, value bpf.MapValue) {
			fmt.Printf(defaultKVFormate, key.String(), value.String())
		}
	}

	if err = m.DumpWithCallback(callback); err != nil {
		klog.Fatalf("dump %s objs failed: %v\n", bpfMapName, err)
	}
}
