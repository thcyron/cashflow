package cf

type Stock struct {
	Name         string
	Symbol       string
	ISIN         string
	Transactions Transactions
}

func (s *Stock) Clone() *Stock {
	cloned := &Stock{}
	*cloned = *s
	cloned.Transactions = s.Transactions.Clone()
	return cloned
}
