package bltfs

type Report struct {
	files struct {
		inTransfer map[uint64]*File
		finished   map[uint64]*File
	}

	bytes struct {
		durable uint64
		total   uint64
	}
}

func (r *Report) Durable() uint64 {
	return r.bytes.durable
}

func (r *Report) Total() uint64 {
	return r.bytes.total
}

func (r *Report) InTransfer() map[uint64]*File {
	return r.files.inTransfer
}

func (r *Report) Finished() map[uint64]*File {
	return r.files.finished
}

type Reporter func(*Report)
