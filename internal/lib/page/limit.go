package page

type Limit struct {
	Want *uint32
	Def  uint32
	Max  uint32
}

func (s *Limit) Limit() uint32 {
	if s.Want != nil {
		return min(*s.Want, s.Max)
	} else {
		if s.Def == 0 {
			panic(ErrDefSize)
		}
		return s.Def
	}
}
