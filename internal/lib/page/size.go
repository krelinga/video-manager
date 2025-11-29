package page

type Sizer struct {
	Want uint32
	Def  uint32
	Max  uint32
}

func (s *Sizer) Size() uint32 {
	if s.Want != 0 {
		return min(s.Want, s.Max)
	} else {
		if s.Def == 0 {
			panic(ErrDefSize)
		}
		return s.Def
	}
}