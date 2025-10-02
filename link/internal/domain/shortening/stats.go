package shortening

type Stats struct {
	activeLinks uint
}

func (s Stats) ActiveLinks() uint { return s.activeLinks }

func (s Stats) HasQuota(quota, need uint) bool {
	return s.activeLinks+need <= quota
}

func NewStats(activeLinks uint) Stats {
	return Stats{activeLinks}
}
