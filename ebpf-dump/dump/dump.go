package dump

import (
	"fmt"
	objects2 "github.com/QQGoblin/ebpf-dump/dump/objects"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/maps/lbmap"
	"github.com/pkg/errors"
	"io"
	"os"
	"text/tabwriter"
)

type CreateDumpCallback func(w io.Writer) func(key bpf.MapKey, value bpf.MapValue)

type Meta struct {
	MapKey      bpf.MapKey
	MapValue    bpf.MapValue
	Parser      bpf.DumpParser
	NewCallback CreateDumpCallback
}

var (
	dumpMeta = map[string]Meta{
		lbmap.Service4MapV2Name:  {MapKey: &lbmap.Service4Key{}, MapValue: &lbmap.Service4Value{}, NewCallback: objects2.NewLBServiceDumpCallback},
		lbmap.Backend4MapV2Name:  {MapKey: &lbmap.Backend4Key{}, MapValue: &lbmap.Backend4Value{}, NewCallback: objects2.NewLBBackendDumpCallback},
		lbmap.RevNat4MapName:     {MapKey: &lbmap.RevNat4Key{}, MapValue: &lbmap.RevNat4Value{}, NewCallback: objects2.NewRevNat4DumpCallback},
		lbmap.SockRevNat4MapName: {MapKey: &objects2.SockRevNat4Key{}, MapValue: &objects2.SockRevNat4Value{}, NewCallback: objects2.NewSockRevNat4DumpCallback},
	}
)

func printHeader(ebpfMapName string) *tabwriter.Writer {

	var w *tabwriter.Writer
	w = tabwriter.NewWriter(os.Stdout, 5, 0, 5, ' ', tabwriter.TabIndent)
	switch ebpfMapName {
	case lbmap.Service4MapV2Name:
		fmt.Fprintln(w, "Frontend\tSlot\tBNum\tBID\tRevNat\tFlags\tFlags2")
	case lbmap.Backend4MapV2Name:
		fmt.Fprintln(w, "ID\tAddress\tFlags\tState")
	case lbmap.RevNat4MapName:
		fmt.Fprintln(w, "RevNat\tOld-Address")
	case lbmap.SockRevNat4MapName:
		fmt.Fprintln(w, "Cookie\tNat-Address\tOld-Address\tRevNat")
	default:
		fmt.Fprintln(w, "Key\tValue")
	}
	return w
}

func defaultCallback(w io.Writer) func(key bpf.MapKey, value bpf.MapValue) {
	return func(key bpf.MapKey, value bpf.MapValue) {
		fmt.Fprintf(w, "%s\t%s\n", key.String(), value.String())
	}
}

func ListMap() error {
	w := tabwriter.NewWriter(os.Stdout, 20, 0, 5, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name")
	for k, _ := range dumpMeta {
		fmt.Fprintln(w, k)
	}
	w.Flush()
	return nil
}

func DumpMap(bpfMapName string) error {

	dm, isOK := dumpMeta[bpfMapName]
	if !isOK {
		return fmt.Errorf("%s dump meta is not registry", bpfMapName)
	}

	m, err := bpf.OpenMap(bpfMapName)
	if err != nil {
		return errors.Wrapf(err, "open %s", bpfMapName)
	}

	m.MapInfo.MapKey = dm.MapKey
	m.MapInfo.MapValue = dm.MapValue
	m.DumpParser = dm.Parser

	if dm.Parser == nil {
		m.DumpParser = bpf.ConvertKeyValue
	}

	w := printHeader(bpfMapName)
	var callback func(key bpf.MapKey, value bpf.MapValue)
	if dm.NewCallback == nil {
		callback = defaultCallback(w)
	} else {
		callback = dm.NewCallback(w)
	}

	if err = m.DumpWithCallback(callback); err != nil {
		return err
	}
	w.Flush()

	return nil
}
