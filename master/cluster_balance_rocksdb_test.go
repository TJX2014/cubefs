package master

import (
	"testing"

	"github.com/cubefs/cubefs/proto"
)

func TestCheckMetaReplicasIsRocksdb(t *testing.T) {
	mp := &MetaPartition{Replicas: []*MetaReplica{{StoreMode: proto.StoreModeMem}, {StoreMode: proto.StoreModeRocksDb}}}
	if !checkMetaReplicasIsRocksdb(mp) {
		t.Fatalf("expected true")
	}
}

func TestIsRocksdbDiskUsageLow(t *testing.T) {
	low := &MetaNode{RocksdbDisks: []*proto.MetaNodeRocksdbInfo{{UsageRatio: gConfig.metaNodeMemLowPer - 0.01}}}
	if !IsRocksdbDiskUsageLow(low) {
		t.Fatalf("expected true for low usage")
	}
	high := &MetaNode{RocksdbDisks: []*proto.MetaNodeRocksdbInfo{{UsageRatio: gConfig.metaNodeMemLowPer + 0.01}}}
	if IsRocksdbDiskUsageLow(high) {
		t.Fatalf("expected false for high usage")
	}
}
