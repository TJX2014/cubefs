package metanode

func (i *Inode) MarshalRocksdb() (result []byte, err error) {
	return i.Marshal()
}

func (i *Inode) UnmarshalRocksdb(raw []byte) (err error) {
	return i.Unmarshal(raw)
}
